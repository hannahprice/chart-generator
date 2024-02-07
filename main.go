package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/snapshot-chromedp/render"
)

type payment struct {
	category string
	cost     float64
}

type paymentGroup struct {
	category string
	cost     int64
}

// Run with: go run main.go --spending-file="spending.csv" --income-file="income.csv"
// Or remove arguments for default file e.g. go run ./
func main() {
	var spendingFileLocation, incomeFileLocation string
	flag.StringVar(&spendingFileLocation, "spending-file", "spending.csv", "File location of spending csv")
	flag.StringVar(&incomeFileLocation, "income-file", "income.csv", "File location of income csv")
	flag.Parse()

	closeSpendingFile, spendingCSV := openFile(spendingFileLocation)
	defer closeSpendingFile()

	payments := parseSpendingCSV(spendingCSV)
	groupedPayments := groupPaymentsByCategory(payments)

	// generate a pie chart image for committed spending
	spendingPie := newPie()
	spendingSeries := spendingPieData(groupedPayments)
	spendingPie.AddSeries("pie", spendingSeries)
	render.MakeChartSnapshot(spendingPie.RenderContent(), "committed-spending.png")

	closeIncomeFile, incomeCSV := openFile(incomeFileLocation)
	defer closeIncomeFile()

	groups := parseIncomeCSV(incomeCSV)

	// generate another pie from static values
	incomePie := newPie()
	incomeSeries := incomePieData(groups)
	incomePie.AddSeries("pie", incomeSeries)
	render.MakeChartSnapshot(incomePie.RenderContent(), "income-breakdown.png")

	// delete the temporary committed-spending.html - it seems to fail to cleanup in the library
	err := os.Remove("committed-spending.html")
	if err != nil {
		log.Fatal("Error removing committed-spending.html")
	}
}

func openFile(fileLocation string) (func() error, [][]string) {
	file, err := os.Open(fileLocation)
	if err != nil {
		log.Fatal("Error opening file")
	}

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Error reading CSV")
	}

	return file.Close, rows
}

func parseSpendingCSV(csv [][]string) []*payment {
	// remove the header row
	csv = csv[1:]

	payments := []*payment{}
	for _, row := range csv {
		cost, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			log.Fatal("Error parsing cost")
		}

		item := &payment{
			category: row[2],
			cost:     cost,
		}
		payments = append(payments, item)
	}
	return payments
}

func groupPaymentsByCategory(payments []*payment) map[string]float64 {
	groupedTotals := map[string]float64{}

	for _, payment := range payments {
		if _, ok := groupedTotals[payment.category]; !ok {
			// category doesn't exist in map
			groupedTotals[payment.category] = payment.cost
			continue
		}

		// category already exists, add to total
		currentTotal := groupedTotals[payment.category]
		groupedTotals[payment.category] = currentTotal + payment.cost
	}

	return groupedTotals
}

func newPie() *charts.Pie {
	pie := charts.NewPie()
	pie.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:           "600px",
			Height:          "500px",
			BackgroundColor: "#FFFFFF",
		}),
		charts.WithAnimation(false),
		charts.WithLegendOpts(opts.Legend{
			Top: "1%",
		}),
	)
	return pie
}

func spendingPieData(groupedPayments map[string]float64) []opts.PieData {
	items := []opts.PieData{}
	for category, total := range groupedPayments {
		nameAndCost := fmt.Sprintf("%s: £%.2f", category, total)
		item := opts.PieData{Name: nameAndCost, Value: total, Label: &opts.Label{Show: opts.Bool(false)}}

		items = append(items, item)
	}
	return items
}

func parseIncomeCSV(csv [][]string) []*paymentGroup {
	groups := []*paymentGroup{}
	for _, row := range csv {
		cost, err := strconv.ParseInt(row[1], 10, 32)
		if err != nil {
			log.Fatal("Error parsing cost")
		}

		item := &paymentGroup{
			category: row[0],
			cost:     cost,
		}
		groups = append(groups, item)
	}
	return groups
}

func incomePieData(groups []*paymentGroup) []opts.PieData {
	items := []opts.PieData{}
	for _, group := range groups {
		nameAndCost := fmt.Sprintf("%s: £%d", group.category, group.cost)
		item := opts.PieData{Name: nameAndCost, Value: group.cost, Label: &opts.Label{Show: opts.Bool(false)}}

		items = append(items, item)
	}
	return items
}
