// Package dot provides an abstraction for painlessly building and printing
// graphviz dot-files
package dot

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Element captures the information of a dot-file element,
// typically corresoponding to one line of the file
type Element interface {
	Write(io.Writer) error
}

// Literal is an element consisting of the corresponding literal string
// printed in the dot-file
type Literal struct {
	Line string
}

// Write writes the literal to a writer
func (lit *Literal) Write(w io.Writer) error {
	_, err := io.WriteString(w, lit.Line)
	return err
}

// VertexDescription is an element containing all the information needed to
// fully describe a dot-file vertex
type VertexDescription struct {
	ID string

	// string attributes
	Label       string
	Group       string
	Color       string
	Style       string
	ColorScheme string
	FontColor   string
	FontName    string
	Shape       string

	// int attributes
	Peripheries int
}

// NewVertexDescription returns a new VertexDescription with the given ID.
func NewVertexDescription(id string) VertexDescription {
	return VertexDescription{
		ID: id,
	}
}

// Write writes the vertex description to a writer
func (v *VertexDescription) Write(w io.Writer) error {
	nodeStr := fmt.Sprintf("%s ", v.ID)
	vertexR := reflect.ValueOf(*v)
	nodeStr += "["
	for i := 1; i < vertexR.NumField(); i++ {
		field := vertexR.Field(i)
		name := strings.ToLower(vertexR.Type().Field(i).Name)

		switch field.Kind() {
		case reflect.String:
			value := field.String()
			if value != "" {
				// for html like tags
				if value[0] == '<' {
					nodeStr += fmt.Sprintf("%s=%s ", name, value)
				} else {
					nodeStr += fmt.Sprintf("%s=\"%s\" ", name, value)
				}
			}
		case reflect.Int:
			value := field.Int()
			if value != 0 {
				nodeStr += fmt.Sprintf("%s=\"%d\" ", name, value)
			}
		}
	}
	nodeStr += "]"
	_, err := io.WriteString(w, nodeStr)
	return err
}

// EdgeDescription is an element containing all the information needed to
// fully describe a dot-file edge
type EdgeDescription struct {
	From     VertexDescription
	To       VertexDescription
	Directed bool

	Style string
}

// Write writes the edge description to a writer
func (e *EdgeDescription) Write(w io.Writer) error {
	var arrow string
	if e.Directed {
		arrow = "->"
	} else {
		arrow = "--"
	}
	edgeStr := fmt.Sprintf("%s %s %s", e.From.ID, arrow, e.To.ID)
	if e.Style != "" {
		edgeStr += fmt.Sprintf(" [ style=\"%s\" ]", e.Style)
	}
	_, err := io.WriteString(w, edgeStr)
	return err
}

// Graph is the graphviz dot-file graph representation.
type Graph struct {
	Name       string
	Body       []Element
	IsSubGraph bool

	// string attributes
	Rank string
}

// NewGraph returns a new dot-file graph object given the provided name
func NewGraph(name string) Graph {
	return Graph{
		Name: name,
	}
}

// AddComment interprets the given argument as the text of a comment and
// schedules the comment to be written in the output dotfile
func (graph *Graph) AddComment(text string) {
	commentStr := fmt.Sprintf("/* %s */", text)
	line := &Literal{
		Line: commentStr,
	}
	graph.Body = append(graph.Body, line)
}

// AddNewLine schedules a newline to be written in the output dotfile
func (graph *Graph) AddNewLine() {
	line := &Literal{
		Line: "", // newline already printed for every line
	}
	graph.Body = append(graph.Body, line)
}

// AddVertex schedules the vertexdescription to be written in the output
// dotfile
func (graph *Graph) AddVertex(v *VertexDescription) {
	graph.Body = append(graph.Body, v)
}

// AddEdge constructs an edgedescription connecting the two vertices given
// as parameters and schedules this element to be written in the output dotfile
func (graph *Graph) AddEdge(v1 *VertexDescription, v2 *VertexDescription, directed bool, style string) {
	edge := &EdgeDescription{
		From:     *v1,
		To:       *v2,
		Directed: directed,
		Style:    style,
	}
	graph.Body = append(graph.Body, edge)
}

// AddSubGraph schedules a newline to be written in the output dotfile.
func (graph *Graph) AddSubGraph(sGraph *Graph) {
	graph.Body = append(graph.Body, sGraph)
}

// WriteDot writes the elements scheduled on this Graph to the provided
// writer to construct a valid dot-file
func (graph *Graph) Write(w io.Writer) error {
	var title string
	if graph.IsSubGraph {
		title = fmt.Sprintf("subgraph %s {\n", graph.Name)
	} else {
		title = fmt.Sprintf("digraph %s {\n", graph.Name)
	}
	_, err := io.WriteString(w, title)
	if err != nil {
		return err
	}

	if graph.Rank != "" {
		_, err = io.WriteString(w, fmt.Sprintf("%s=\"%s\"\n", "rank", graph.Rank))
		if err != nil {
			return err
		}
	}

	for _, line := range graph.Body {
		err = line.Write(w)
		_, err2 := io.WriteString(w, "\n")
		if err != nil || err2 != nil {
			return err
		}

	}

	_, err = io.WriteString(w, "}")
	return err
}
