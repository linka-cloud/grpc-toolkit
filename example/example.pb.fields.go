package main

var HelloRequestFields = struct {
	Name string
}{
	Name: "name",
}

var HelloReplyFields = struct {
	Message string
}{
	Message: "message",
}

var HelloStreamRequestFields = struct {
	Name  string
	Count string
}{
	Name:  "name",
	Count: "count",
}
