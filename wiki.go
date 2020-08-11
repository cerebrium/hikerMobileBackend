package main

import (
	"fmt"
	"io/ioutil"
)

type Page struct {
	// structure for the whole page
	Title string
	Body  []byte
}

func (p *Page) save() error {
	// := is a decleration of variable scoped to the function
	filename := p.Title + ".txt"

	// the octal makes it so the read/write privleges are only for this user
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func main() {
	p1 :=
		&Page{Title: "TestPage",
			Body: []byte("This is a sample Page.")}
	p1.save()
	p2, _ := loadPage("TestPage")
	fmt.Println(string(p2.Body))
}
