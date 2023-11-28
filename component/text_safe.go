package component

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"

	filter "github.com/antlinker/go-dirtyfilter"
	"github.com/antlinker/go-dirtyfilter/store"
)

type TextSafe struct {
	filter *filter.DirtyManager
}

func (s *TextSafe) NewFilter() error {
	fi, err := os.Open("config/words_filter.txt")
	if err != nil {
		log.Printf("open words_filter err %v", err)
		return err
	}

	defer func() {
		_ = fi.Close()
	}()

	words := []string{}
	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		words = append(words, string(a))
	}

	memStore, err := store.NewMemoryStore(store.MemoryConfig{
		DataSource: words,
	})

	if err != nil {
		log.Printf("NewMemoryStore err %v", err)
		return err
	}

	s.filter = filter.NewDirtyManager(memStore)
	return nil
}

func (s *TextSafe) Filter(filterText string) string {
	result, err := s.filter.Filter().Filter(filterText, '*', '@')
	if err != nil {
		log.Print(err)
		return ""
	}

	if result != nil {
		for _, w := range result {
			filterText = strings.ReplaceAll(filterText, w, "*")
		}
	}

	return filterText
}
