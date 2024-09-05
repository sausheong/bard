package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/hako/durafmt"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"golang.org/x/exp/rand"
)

var cyan = color.New(color.FgCyan).SprintFunc()
var yellow = color.New(color.FgHiYellow).SprintFunc()

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	if _, err := os.Stat("md"); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir("md", os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	if _, err := os.Stat("plots"); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir("plots", os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	if _, err := os.Stat("html"); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir("html", os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
}

func main() {
	app := &cli.App{
		Name: "Bard",
		Authors: []*cli.Author{
			{
				Name:  "Chang Sau Sheong",
				Email: "sausheong@gmail.com",
			},
		},
		Copyright: "(c) 2024 Chang Sau Sheong",
		Usage:     "Using AI to create stories",
		Commands: []*cli.Command{
			{
				Name:  "prepare",
				Usage: "prepare a plot for the story",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "seedfile",
						Value:    "seed.txt",
						Usage:    "seed file to use",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "plotfile",
						Value:    "plot.txt",
						Usage:    "plot file to create",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					t0 := time.Now()
					plot(c.String("seedfile"), c.String("plotfile"), c.String("model"))
					elapsed := durafmt.Parse(time.Since(t0)).LimitFirstN(1)
					fmt.Println(yellow(fmt.Sprintf("Plot %s generated in %s",
						c.String("plotfile"), elapsed)))
					return nil
				},
			},
			{
				Name:  "generate",
				Usage: "generate a story",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "num_parts",
						Aliases: []string{"n"},
						Value:   4,
						Usage:   "number of parts, must be more than 3",
						Action: func(ctx *cli.Context, v int) error {
							if v <= 3 {
								return fmt.Errorf("you have tried to set %d parts. "+
									"Each story must have at least 4 parts", v)
							}
							return nil
						},
					},
					&cli.StringFlag{
						Name:     "plotfile",
						Value:    "plot.txt",
						Usage:    "plot file to use",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "mdfile",
						Value:    "story.md",
						Usage:    "markdown file to convert",
						Required: true,
					},
					&cli.BoolFlag{
						Name:     "verbose",
						Aliases:  []string{"v"},
						Value:    false,
						Usage:    "print parts to screen",
						Required: false,
					},
				},
				Action: func(c *cli.Context) error {
					t0 := time.Now()
					story(c.String("plotfile"),
						c.String("mdfile"),
						c.String("model"),
						c.Int("num_parts"),
						c.Bool("verbose"))
					elapsed := durafmt.Parse(time.Since(t0)).LimitFirstN(1)
					fmt.Println(yellow(fmt.Sprintf("Story %s generated in %s",
						c.String("mdfile"), elapsed)))
					return nil

				},
			},
			{
				Name:  "convert",
				Usage: "convert the markdown file to HTML",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "mdfile",
						Value:    "story.md",
						Usage:    "markdown file to convert",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "htmlfile",
						Value:    "output.html",
						Usage:    "output HTML file name",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					startTime := time.Now()
					convert(c.String("mdfile"), c.String("htmlfile"))
					fmt.Println(yellow(fmt.Sprintf(
						"HTML file %s converted in %s",
						c.String("htmlfile"), time.Since(startTime))))
					return nil
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "model",
				Aliases: []string{"m"},
				Value:   "llama3.1",
				Usage:   "large language model to use",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// generate the plot, given a seed file
func plot(seedfile string, plotfile string, model string) {
	fmt.Println(yellow("Preparing the plot from the seed now."))
	plotSeed, err := readFile(seedfile)
	if err != nil {
		log.Fatalln("Cannot read seed file:", err)
	}
	prompt := fmt.Sprintf(prepare, plotSeed)
	plot, err := generate(model, prompt, randomInt())
	if err != nil {
		log.Fatalln("Cannot generate plot:", err)
	}

	fmt.Println(yellow("> This is the plot used in the story."))
	fmt.Println()
	fmt.Println(cyan(plot))
	fmt.Println()
	saveFile(plotfile, plot, true)
}

// generate the parts, given the plot file
func story(plotfile string, mdfile string, model string, numParts int, verbose bool) {
	fmt.Println(yellow("Generating story from plot file now."))
	var prompt string
	var draft string
	seed := randomInt()

	plot, err := readFile(plotfile)
	if err != nil {
		log.Fatalln("Cannot read plot file:", err)
	}

	// 1. generate the first part
	fmt.Println(yellow("> part 1"))
	prompt = "[overall plot]\n" + plot + "\n---\n" + first
	firstPart, err := generate(model, prompt, seed)
	if err != nil {
		log.Fatalln("Cannot generate first part:", err)
	}
	if verbose {
		fmt.Println()
		fmt.Println(firstPart)
		fmt.Println()
	}
	draft += "\n\n" + firstPart
	i := 1
	// 2. generate the next few parts
	for ; i < numParts-1; i++ {
		fmt.Println(yellow(fmt.Sprintf("> part %d", i+1)))
		prompt = "[overall plot]\n" + plot + "\n---\n" + "[story so far]\n" + draft + "\n---\n" + next
		nextPart, err := generate(model, prompt, seed)
		if err != nil {
			log.Fatalln("Cannot generate next part:", err)
		}
		if verbose {
			fmt.Println()
			fmt.Println(nextPart)
			fmt.Println()
		}
		draft += "\n\n" + nextPart
	}

	// 3. generate the final part
	fmt.Println(yellow(fmt.Sprintf("> part %d (final part)", i+1)))
	prompt = "[overall plot]\n" + plot + "\n---\n" + "[story so far]\n" + draft + "\n---\n" + last
	lastPart, _ := generate(model, prompt, seed)
	if verbose {
		fmt.Println()
		fmt.Println(lastPart)
		fmt.Println()
	}
	draft += "\n\n" + lastPart

	// save the draft
	saveFile(mdfile, draft, true)
}

// convert the markdown to html
func convert(mdfile string, htmlfile string) {
	// read the markdown file
	md, err := readFile(mdfile)
	if err != nil {
		log.Fatalln("Cannot read markdown file:", err)
	}
	var buf bytes.Buffer
	// create the markdown processor
	markdown := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	// convert the markdown to html
	err = markdown.Convert([]byte(md), &buf)
	if err != nil {
		log.Fatalln("Cannot convert markdown to pdf:", err)
	}

	// read the template
	template, err := os.ReadFile("output.template")
	if err != nil {
		log.Fatalln("Cannot read template file:", err)
	}

	// save the html
	output := fmt.Sprintf(string(template), buf.String())
	err = os.WriteFile(htmlfile, []byte(output), 0644)
	if err != nil {
		log.Fatalln("Cannot save pdf file:", err)
	}
}

// read a file and return the contents as a string
func readFile(filepath string) (string, error) {
	ba, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalln("Cannot read file:", err)
		return "", err
	}
	return string(ba), nil
}

// save the contents of the string to a file
func saveFile(filepath string, content string, overwrite bool) error {
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln("Cannot save file:", err)
		return err
	}
	defer file.Close()
	if overwrite {
		file.Truncate(0)
	}
	file.WriteString(content)
	return nil
}

func randomInt() int {
	rand.Seed(uint64(time.Now().UnixNano()))
	return rand.Intn(math.MaxInt)
}
