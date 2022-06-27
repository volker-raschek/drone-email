package domain

type Commit struct {
	Author  *Author
	Branch  string
	Link    string
	Message string
	Ref     string
	Sha     string
}
