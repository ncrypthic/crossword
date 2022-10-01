package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	DIRECTION_DOWN  = 1
	DIRECTION_RIGHT = 2
)

var (
	showResult bool
	matrixSize int
	addr       string
	runAsCli   bool
	wordlist   string

	hints    map[string]string = make(map[string]string)
	words    []string          = make([]string, 0)
	wordsLen map[string]int    = make(map[string]int)
	lenWords map[int][]string  = make(map[int][]string)
)

type Position struct {
	Y, X int
}

func main() {
	flag.BoolVar(&showResult, "r", false, "print result")
	flag.BoolVar(&runAsCli, "cli", false, "run as command-line")
	flag.IntVar(&matrixSize, "s", 8, "matrix size")
	flag.StringVar(&addr, "p", "0.0.0.0:8088", "http server port")
	flag.StringVar(&wordlist, "l", "", "word list path")
	flag.Parse()
	if wordlist == "" {
		fmt.Println("error: word list is required")
		fmt.Println()
		flag.PrintDefaults()
		os.Exit(1)
	}
	var r *bufio.Reader
	switch {
	case len(os.Args) == 0:
		r = bufio.NewReader(os.Stdin)
	default:
		f, err := os.OpenFile(wordlist, os.O_RDONLY, 0x664)
		if err != nil {
			panic(err)
		}
		r = bufio.NewReader(f)
	}
	for {
		b, _, err := r.ReadLine()
		if errors.Is(err, io.EOF) {
			break
		}
		str := string(b)
		segments := strings.Split(str, ",")
		w := strings.ToUpper(segments[0])
		hints[w] = strings.Trim(segments[1], " ")
		l := len(b)
		words = append(words, w)
		wordsLen[w] = l
		if _, ok := lenWords[l]; !ok {
			lenWords[l] = make([]string, 0)
		}
		lenWords[l] = append(lenWords[l], w)
	}
	if runAsCli {
		printCli(matrixSize, words)
		return
	}
	serveHttp(addr, words)
}

func printCli(size int, words []string) {
	m, res := createMatrix(matrixSize, words)
	resHints := make([]string, 0)
	for w := range res {
		resHints = append(resHints, hints[w])
	}
	fmt.Println("=======================================")
	printMatrix(m)
	fmt.Println("=======================================")
	printHints(resHints)
	if showResult {
		printResult(res)
	}
}

// 1. create a square matrix of certain size
// 2. pick a random position to begin
// 2.1 count length to bottom
// 2.1.1 check intersection for other letters on the path
// 2.1.2 get possible words to match length to bottom & intersections
// 2.2 count length to right
// 2.2.1 check intersection for other letters on the path
// 2.2.2 get possible words to match length to bottom & intersections
// 3. foreach empty element in matrix, insert a random char
// 4. print the matrix
func createMatrix(s int, words []string) ([][]byte, map[string]Position) {
	rand.Seed(time.Now().Unix())
	res := make(map[string]Position)
	m := make([][]byte, s)
	for i := 0; i < s; i++ {
		m[i] = make([]byte, s)
	}
	visited := make([]Position, 0)
	for len(visited) < s^2 {
		p := Position{
			Y: rand.Int() % (s - 1),
			X: rand.Int() % (s - 1),
		}
		if w, ok := fillDown(m, p, words); ok {
			res[w] = p
			words = removeWord(words, w)
			continue
		}
		if w, ok := fillRight(m, p, words); ok {
			res[w] = p
			words = removeWord(words, w)
			continue
		}
		visited = append(visited, p)
	}

	chars := []rune{
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H',
		'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P',
		'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X',
		'Y', 'Z',
	}

	for i := 0; i < s; i++ {
		for j := 0; j < s; j++ {
			if m[i][j] != 0 {
				continue
			}
			r := rand.Int() % 25
			m[i][j] = byte(chars[r])
		}
	}
	return m, res
}

func removeWord(words []string, word string) []string {
	res := make([]string, 0)
	for _, w := range words {
		if w == word {
			continue
		}
		res = append(res, w)
	}
	return res
}

func fillDown(m [][]byte, p Position, words []string) (string, bool) {
	c := len(m) - p.X
	possibleWords := getPossibleWords(m, p, words, c)
	return chooseWord(m, possibleWords, p, DIRECTION_DOWN, c)
}

func fillRight(m [][]byte, p Position, words []string) (string, bool) {
	c := len(m) - p.Y
	possibleWords := getPossibleWords(m, p, words, c)
	return chooseWord(m, possibleWords, p, DIRECTION_RIGHT, c)
}

func getPossibleWords(m [][]byte, pos Position, words []string, count int) []string {
	res := make([]string, 0)
	for _, w := range words {
		if len(w) <= count {
			res = append(res, w)
		}
	}
	return res
}

func getIntersections(m [][]byte, pos Position, direction, count int) []byte {
	intersections := make([]byte, count)
	switch direction {
	case DIRECTION_RIGHT:
		x := pos.Y
		i := 0
		for i < count-1 {
			c := m[x][pos.X]
			if c != 0 {
				intersections[i] = c
			}
			i++
			x++
		}
	default:
		y := pos.X
		i := 0
		for i < count-1 {
			c := m[pos.Y][y]
			if c != 0 {
				intersections[i] = c
			}
			i++
			y++
		}
	}
	return intersections
}

func chooseWord(m [][]byte, words []string, pos Position, dir, count int) (string, bool) {
WordsLoop:
	for _, w := range words {
		if len(w) > count {
			continue
		}
		intersections := getIntersections(m, pos, dir, count)
	CheckIntersection:
		for i, c := range intersections {
			if c == 0 {
				continue CheckIntersection
			}
			if len(w) > i && w[i] != c {
				continue WordsLoop
			}
		}
		switch dir {
		case DIRECTION_RIGHT:
			for i, r := range w {
				m[pos.Y+i][pos.X] = byte(r)
			}
		default:
			for i, r := range w {
				m[pos.Y][pos.X+i] = byte(r)
			}
		}
		return w, true
	}
	return "", false
}

func printMatrix(m [][]byte) {
	fmt.Printf("   ")
	for i := 0; i < len(m); i++ {
		fmt.Printf("%2d ", i+1)
	}
	y := 1
	for _, x := range m {
		fmt.Println()
		fmt.Printf("%2d ", y)
		for _, y := range x {
			fmt.Printf("%2s ", string(y))
		}
		fmt.Println()
		y++
	}
}

func printHints(hints []string) {
	for i, h := range hints {
		fmt.Print(i+1, " :")
		fmt.Println(h)
	}
}

func printResult(res map[string]Position) {
	for w, p := range res {
		fmt.Printf("%s, X: %d, Y: %d\n", w, p.X+1, p.Y+1)
	}
}

type CrosswordRequest struct {
	Size int `json:"size,default:10"`
}

type CrosswordResponse struct {
	Matrix [][]string          `json:"matrix"`
	Answer map[string]Position `json:"ans"`
	Hints  map[string]string   `json:"hints"`
}

type HTTPResponse struct {
	w http.ResponseWriter
}

func (h *HTTPResponse) Send(code int, data interface{}) {
	h.w.WriteHeader(code)
	json.NewEncoder(h.w).Encode(data)
}

func Response(w http.ResponseWriter) *HTTPResponse {
	return &HTTPResponse{w}
}

func serveHttp(port string, words []string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/"+r.URL.Path)
	})
	http.HandleFunc("/crossword", func(w http.ResponseWriter, r *http.Request) {
		res := Response(w)
		if r.Method != http.MethodPost {
			res.Send(http.StatusMethodNotAllowed, nil)
			return
		}
		var payload CrosswordRequest
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			res.Send(http.StatusBadRequest, err)
			return
		}
		matrix, answer := createMatrix(payload.Size, words)
		crossword := make([][]string, len(matrix))
		for x, row := range matrix {
			crossword[x] = make([]string, len(matrix))
			for y, cell := range row {
				crossword[x][y] = string(cell)
			}
		}
		resHints := make(map[string]string)
		for w := range answer {
			resHints[w] = hints[w]
		}
		res.Send(http.StatusOK, &CrosswordResponse{
			Matrix: crossword,
			Answer: answer,
			Hints:  resHints,
		})
	})
	fmt.Println("Crossword Puzzle running at", addr)
	http.ListenAndServe(addr, nil)
}
