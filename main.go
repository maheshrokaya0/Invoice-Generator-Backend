package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type InvoiceData struct {
	InvoiceNumber      int64
	IssuedDate         string
	DueDate            string
	ClientName         string
	ClientAddress      string
	ClientCityStateZip string
	YourName           string
	YourAddress        string
	YourCityStateZip   string

	Rows     []RowData
	SubTotal float64
	Discount float64
	Tax      float64
	Total    float64
	Note     string
}

type RowData struct {
	Name     string
	Quantity int64
	Rate     float64
	Amount   float64
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/api/generate-invoice", generateInvoice).Methods(http.MethodPost, http.MethodOptions)
	router.Use(mux.CORSMethodMiddleware(router))

	fmt.Println("Server runnig at port 3000")
	http.ListenAndServe(":3000", router)
}

func generateInvoice(w http.ResponseWriter, r *http.Request) {

	enableCors(&w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	fileID := uuid.NewString()

	var invoiceData InvoiceData

	err := json.NewDecoder(r.Body).Decode(&invoiceData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current working directory:", err)
		return
	}

	pdf := fpdf.New("P", "mm", "A4", filepath.Join(currentDir, "assets", "fonts"))
	pdf.AddUTF8Font("Outfit", "", "Outfit-Regular.ttf")
	pdf.AddUTF8Font("Outfit", "B", "Outfit-SemiBold.ttf")

	pdf.SetMargins(5, 10, 5)
	pdf.SetCellMargin(5)
	pdf.AddPage()

	pdf.SetFont("Outfit", "", 28)

	pdf.SetTextColor(50, 60, 50)
	pdf.Cell(100, 10, fmt.Sprintf("INVOICE #%d", invoiceData.InvoiceNumber))
	pdf.Ln(10)

	pdf.SetFont("Outfit", "B", 12)
	if invoiceData.IssuedDate != "" {
		pdf.Cell(100, 10, fmt.Sprintf("Issued : %s", invoiceData.IssuedDate))
	} else {
		pdf.Cell(100, 10, fmt.Sprintf("Issued : %s", time.Now().Format("2006-01-02")))
	}
	pdf.Ln(8)

	if invoiceData.DueDate != "" {
		pdf.Cell(100, 10, fmt.Sprintf("Due : %s", invoiceData.DueDate))
	} else {
		pdf.Cell(100, 10, fmt.Sprintf("Due : On Receipt"))
	}
	pdf.Ln(20)

	pdf.Cell(100, 10, "BILL TO :")
	pdf.Cell(100, 10, "PAY TO :")
	pdf.Ln(8)

	pdf.SetFont("Outfit", "", 12)
	pdf.Cell(100, 10, fmt.Sprintf("%s", invoiceData.ClientName))
	pdf.Cell(100, 10, fmt.Sprintf("%s", invoiceData.YourName))
	pdf.Ln(8)

	pdf.Cell(100, 10, fmt.Sprintf("%s", invoiceData.ClientAddress))
	pdf.Cell(100, 10, fmt.Sprintf("%s", invoiceData.YourAddress))
	pdf.Ln(8)

	pdf.Cell(100, 10, fmt.Sprintf("%s", invoiceData.ClientCityStateZip))
	pdf.Cell(100, 10, fmt.Sprintf("%s", invoiceData.YourCityStateZip))
	pdf.Ln(24)

	pdf.SetFont("Outfit", "B", 12)
	pdf.SetFillColor(0, 0, 0)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(90, 10, "Items", "", 0, "", true, 0, "")
	pdf.CellFormat(30, 10, "Qty", "", 0, "", true, 0, "")
	pdf.CellFormat(40, 10, "Rate", "", 0, "", true, 0, "")
	pdf.CellFormat(40, 10, "Amount", "", 0, "", true, 0, "")
	pdf.Ln(10)

	pdf.SetFont("Outfit", "", 12)
	pdf.SetTextColor(0, 0, 0)

	for _, row := range invoiceData.Rows {
		addInvoiceItem(pdf, row.Name, row.Quantity, row.Rate, row.Amount)
	}

	pdf.Ln(10)
	pdf.Cell(120, 10, "")

	pdf.SetFont("Outfit", "B", 12)
	pdf.CellFormat(40, 10, "SubTotal", "", 0, "", false, 0, "")
	pdf.SetFont("Outfit", "", 12)
	pdf.CellFormat(40, 10, fmt.Sprintf("$%.2f", invoiceData.SubTotal), "", 0, "", false, 0, "")
	pdf.Ln(10)

	if invoiceData.Discount != 0 {
		pdf.Cell(120, 10, "")
		pdf.SetFont("Outfit", "B", 12)
		pdf.CellFormat(40, 10, "Discount", "", 0, "", false, 0, "")
		pdf.SetFont("Outfit", "", 12)
		pdf.CellFormat(40, 10, fmt.Sprintf("%.2f %%", invoiceData.Discount), "", 0, "", false, 0, "")
		pdf.Ln(10)
	}

	if invoiceData.Tax != 0 {
		pdf.Cell(120, 10, "")
		pdf.SetFont("Outfit", "B", 12)
		pdf.CellFormat(40, 10, "Tax", "", 0, "", false, 0, "")
		pdf.SetFont("Outfit", "", 12)
		pdf.CellFormat(40, 10, fmt.Sprintf("%.2f %%", invoiceData.Tax), "", 0, "", false, 0, "")
		pdf.Ln(10)
	}

	pdf.Cell(120, 10, "")
	pdf.SetFont("Outfit", "B", 12)
	pdf.CellFormat(40, 10, "Total", "", 0, "", false, 0, "")
	pdf.SetFont("Outfit", "", 12)
	pdf.CellFormat(40, 10, fmt.Sprintf("$%.2f", invoiceData.Total), "", 0, "", false, 0, "")

	pdf.Ln(20)
	pdf.SetFont("Outfit", "B", 12)
	pdf.Cell(20, 10, "Notes: ")
	pdf.SetFont("Outfit", "", 12)
	pdf.Cell(180, 10, fmt.Sprintf("%s", invoiceData.Note))
	fmt.Println(invoiceData.Note)

	err = pdf.OutputFileAndClose(filepath.Join(currentDir, fileID+".pdf"))
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Invoice created successfully")
	}

	pdfPath := filepath.Join(fileID + ".pdf")

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=invoice.pdf")

	http.ServeFile(w, r, pdfPath)

	err = os.Remove(pdfPath)
	if err != nil {
		fmt.Println("Error deleting PDF file:", err)
	} else {
		fmt.Println("Invoice file deleted successfully")
	}
}

func addInvoiceItem(pdf *fpdf.Fpdf, item string, quantity int64, rate float64, amount float64) {
	pdf.Cell(90, 10, item)
	pdf.Cell(30, 10, fmt.Sprintf("%d", quantity))
	pdf.Cell(40, 10, fmt.Sprintf("$%.2f", rate))
	pdf.Cell(40, 10, fmt.Sprintf("$%.2f", amount))
	pdf.Ln(10)
}
