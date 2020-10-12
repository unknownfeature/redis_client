package console

type History struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type Request struct {
	Db      int    `json:"db"`
	Command string `json:"command"`
}

type Response struct {
	History History `json:"history"`
	Prompt  string  `json:"prompt"`
	Db      int     `json:"db"`
}

