package model

type TableNames struct {
	Tables []string `json:"names"`
}

func NewTableNames() *TableNames {
	return &TableNames{
		Tables: []string{
			"m_xxx",
		},
	}
}