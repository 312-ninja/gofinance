/*
Copyright 2016, Matthias Fluor

This is a simple budgeting web app, based on the blog-post by Alex Recker:
https://alexrecker.com/our-new-sid-meiers-civilization-inspired-budget.html

*/

package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// Make the DB global for all
var db *sql.DB

func handleEdit(w http.ResponseWriter, r *http.Request, pr httprouter.Params) {
	t, _ := template.ParseFiles("templates/edit.html")
	entryID := pr.ByName("id")
	var entry int
	entry, _ = strconv.Atoi(entryID)
	trans := getSingle(db, entry, pr.ByName("type"))
	var fixcheck bool
	if pr.ByName("type") == "fixed" {
		fixcheck = true
	}
	t.Execute(w, map[string]interface{}{"trans": trans, "transtype": pr.ByName("type"), "fixcheck": fixcheck})

}
func editEntry(w http.ResponseWriter, r *http.Request, pr httprouter.Params) {
	r.ParseForm()
	income := false
	description := r.Form["description"][0]
	amountstr := r.Form["amount"][0]
	amount, erra := strconv.ParseFloat(amountstr, 64)
	if erra != nil {
		panic(erra)
	}
	incomecheck := r.Form["income"]
	if len(incomecheck) == 0 {
		income = false
	} else {
		income = true
	}
	recurrence := ""
	if pr.ByName("type") == "fixed" {
		recurrence = strings.ToLower(r.Form["recurrence"][0])
	}
	idstr := pr.ByName("id")
	idint, _ := strconv.Atoi(idstr)
	ChangeItem(db, Transaction{ID: idint, Description: description, Amount: amount, Income: income, Recurrence: recurrence}, pr.ByName("type"))
	// Get back to the main page
	http.Redirect(w, r, "/", 301)
}

func getInput(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm()
	income := false
	description := r.Form["description"][0]
	amountstr := r.Form["amount"][0]
	amount, erra := strconv.ParseFloat(amountstr, 64)
	if erra != nil {
		panic(erra)
	}
	incomecheck := r.Form["income"]
	if len(incomecheck) == 0 {
		income = false
	} else {
		income = true
	}
	StoreItem(db, Transaction{Description: description, Amount: amount, Income: income}, "transaction")
	// Get back to the main page
	http.Redirect(w, r, "/", 301)
}

func getFixInput(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm()
	income := false
	description := r.Form["description"][0]
	amountstr := r.Form["amount"][0]
	recurrence := r.Form["recurrence"][0]
	amount, erra := strconv.ParseFloat(amountstr, 64)
	if erra != nil {
		panic(erra)
	}
	incomecheck := r.Form["income"]
	if len(incomecheck) == 0 {
		income = false
	} else {
		income = true
	}
	influence := calcRate(Transaction{Recurrence: recurrence, Amount: amount, Income: income})
	StoreItem(db, Transaction{Description: description, Amount: amount, Income: income, Recurrence: recurrence, Influence: influence}, "fixed")
	// Get back to the main page
	http.Redirect(w, r, "/", 301)
}

// Handler to display the main page - with db-values
func renderMain(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		panic(err)
	}
	// Read the Database to get the current stuff (Date = today)
	fixed := ReadItem(db, "fixed")
	trans := ReadItem(db, "transaction")
	smallerThanZero := false
	magicNumber := baseMagic(db)
	currentNumber := currentMagic(db)
	if currentNumber <= 0 {
		smallerThanZero = true
	}
	t.Execute(w, map[string]interface{}{"fix": fixed, "tran": trans, "mn": magicNumber, "curr": currentNumber, "check": smallerThanZero})
}

// Handler for the insertion
func renderInsert(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, err := template.ParseFiles("templates/input.html")
	if err != nil {
		panic(err)
	}
	t.Execute(w, "")
}

// Handler for the insertion
func renderNewFix(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, err := template.ParseFiles("templates/inputfix.html")
	if err != nil {
		panic(err)
	}
	t.Execute(w, "")
}

func main() {
	// Creates the table on first run if it doesn't exist
	const dbpath = "gofin.db"
	db = initDB(dbpath)
	defer db.Close()
	CreateTable(db)
	// Handlers
	router := httprouter.New()
	router.GET("/", renderMain)
	router.GET("/new/transaction", renderInsert)
	router.GET("/new/fixed", renderNewFix)
	router.GET("/edit/:type/:id", handleEdit)
	router.POST("/confirm/new/transaction", getInput)
	router.POST("/confirm/edit/:type/:id", editEntry)
	router.POST("/confirm/new/fixed", getFixInput)
	// Start the Webserver
	err := http.ListenAndServe(":8080", router) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", router)
	}
}