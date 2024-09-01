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
				Name:    "prepare",
				Aliases: []string{"p"},
				Usage:   "prepare a plot for the story",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "seedfile",
						Aliases:  []string{"s"},
						Value:    "seed.txt",
						Usage:    "seed file to use",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					startTime := time.Now()
					seed := randomInt()
					title := c.String("title")
					model := c.String("model")
					seedfile := c.String("seedfile")
					generatePlot(seedfile, seed, title, model)
					fmt.Println(yellow(fmt.Sprintf("Done in %s", time.Since(startTime))))
					return nil
				},
			},
			{
				Name:    "generate",
				Aliases: []string{"g"},
				Usage:   "generate a story",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "num_chapters",
						Aliases: []string{"n"},
						Value:   4,
						Usage:   "number of chapters, must be more than 3",
						Action: func(ctx *cli.Context, v int) error {
							if v <= 3 {
								return fmt.Errorf("you have tried to set %d chapters. "+
									"Each story must have at least 4 chapters", v)
							}
							return nil
						},
					},
					&cli.StringFlag{
						Name:     "plotfile",
						Aliases:  []string{"p"},
						Value:    "",
						Usage:    "plot file to use",
						Required: true,
					},
					&cli.BoolFlag{
						Name:     "verbose",
						Aliases:  []string{"v"},
						Value:    false,
						Usage:    "print chapters to screen",
						Required: false,
					},
				},
				Action: func(c *cli.Context) error {
					startTime := time.Now()
					seed := randomInt()
					title := c.String("title")
					model := c.String("model")
					numChapters := c.Int("num_chapters")
					plotfile := c.String("plotfile")
					verbose := c.Bool("verbose")
					generateChapters(plotfile, seed, title, model, numChapters, verbose)
					fmt.Println(yellow(fmt.Sprintf("Done in %s", time.Since(startTime))))
					return nil

				},
			},
			{
				Name:    "convert",
				Aliases: []string{"c"},
				Usage:   "convert the markdown file to HTML",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "mdfile",
						Aliases:  []string{"m"},
						Value:    "story.md",
						Usage:    "markdown file to convert",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "outputfile",
						Aliases:  []string{"o"},
						Value:    "story.html",
						Usage:    "output HTML file name",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					startTime := time.Now()
					convertToHtml(c.String("mdfile"), c.String("outputfile"))
					fmt.Println(yellow(fmt.Sprintf("Done in %s", time.Since(startTime))))
					return nil
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "title",
				Aliases: []string{"t"},
				Value:   "My AI Generated Story",
				Usage:   "the title of the story to create",
			},
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
func generatePlot(seedfile string, seed int, title string, model string) {
	fmt.Println(yellow("Preparing the plot from the seed now."))
	startTime := time.Now()
	plotSeed, err := readFile(seedfile)
	if err != nil {
		log.Fatalln("Cannot read seed file:", err)
	}
	prompt := fmt.Sprintf(prepare, plotSeed)
	plot, err := generate(model, prompt, seed)
	if err != nil {
		log.Fatalln("Cannot generate plot:", err)
	}

	fmt.Println(yellow("> This is plot used in the story."))
	fmt.Println()
	fmt.Println(cyan(plot))
	fmt.Println()
	saveFile(fmt.Sprintf("plots/%s.plot", title), plot, true)
	fmt.Println(yellow(fmt.Sprintf("Plot generated in %s, saved to plots/%s.plot", time.Since(startTime), title)))
}

// generate the chapters, given the plot file
func generateChapters(plotfile string, seed int, title string, model string, numChapters int, verbose bool) {

	fmt.Println(yellow("Generating chapters from plot file now."))
	var prompt string
	var draft string

	plot, err := readFile(plotfile)
	if err != nil {
		log.Fatalln("Cannot read plot file:", err)
	}

	// 1. generate the first chapter
	fmt.Println(yellow("> chapter 1"))
	prompt = "[overall plot]\n" + plot + "\n---\n" + first
	firstChapter, err := generate(model, prompt, seed)
	if err != nil {
		log.Fatalln("Cannot generate first chapter:", err)
	}
	if verbose {
		fmt.Println()
		fmt.Println(firstChapter)
		fmt.Println()
	}
	draft += "\n\n" + firstChapter

	// 2. generate the next chapter
	fmt.Println(yellow("> chapter 2"))
	prompt = "[overall plot]\n" + plot + "\n---\n" + "[story so far]\n" + firstChapter + "\n---\n" + next
	nextChapter, err := generate(model, prompt, seed)
	if err != nil {
		log.Fatalln("Cannot generate next chapter:", err)
	}
	if verbose {
		fmt.Println()
		fmt.Println(nextChapter)
		fmt.Println()
	}
	draft += "\n\n" + nextChapter

	// 3. generate the next few chapters
	for i := 0; i < numChapters-2; i++ {
		fmt.Println(yellow(fmt.Sprintf("> chapter %d", i+3)))
		prompt = "[overall plot]\n" + plot + "\n---\n" + "[story so far]\n" + draft + "\n---\n" + next
		nextChapter, err = generate(model, prompt, seed)
		if err != nil {
			log.Fatalln("Cannot generate next chapter:", err)
		}
		if verbose {
			fmt.Println()
			fmt.Println(nextChapter)
			fmt.Println()
		}
		draft += "\n\n" + nextChapter
	}

	// 4. generate the final chapter
	fmt.Println(yellow("> final chapter!"))
	prompt = "[overall plot]\n" + plot + "\n---\n" + "[story so far]\n" + draft + "\n---\n" + last
	lastChapter, _ := generate(model, prompt, seed)
	if verbose {
		fmt.Println()
		fmt.Println(lastChapter)
		fmt.Println()
	}
	draft += "\n\n" + lastChapter

	// save the draft
	saveFile(fmt.Sprintf("md/%s.md", title), draft, true)
}

// convert the markdown to html
func convertToHtml(mdfile string, outputfile string) {
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
	err = os.WriteFile(fmt.Sprintf("html/%s", outputfile), []byte(output), 0644)
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
