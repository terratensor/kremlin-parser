package main

import (
	"github.com/terratensor/kremlin-parser/internal/config"
	"github.com/terratensor/kremlin-parser/internal/parser"
	"log"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	cfg := config.MustLoad()

	//var pageCount, outputPath string
	//
	//flag.StringVarP(&cfg.Parser.OutputPath, "output", "o", "./data", "путь сохранения файлов")
	//flag.StringVarP(&cfg.Parser.PageCount, "page-count", "p", "1", "спарсить указанное количество страниц")
	//flag.Parse()

	prs := parser.New(cfg)
	prs.Parse()
	log.Printf("all pages were successfully parsed")

}
