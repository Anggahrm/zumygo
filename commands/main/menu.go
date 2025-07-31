package commands

import (
	"fmt"
	"sort"
	"zumygo/helpers"
	"zumygo/libs"
)

type item struct {
	Name     []string
	IsPrefix bool
}

type tagSlice []string

func (t tagSlice) Len() int {
	return len(t)
}

func (t tagSlice) Less(i int, j int) bool {
	return t[i] < t[j]
}

func (t tagSlice) Swap(i int, j int) {
	t[i], t[j] = t[j], t[i]
}

func menu(conn *libs.IClient, m *libs.IMessage) bool {
	var str string
	str += fmt.Sprintf("Nah %s, Ini List Command nya\n\n", m.Info.PushName)
	var tags map[string][]item
	for _, list := range libs.GetList() {
		if tags == nil {
			tags = make(map[string][]item)
		}
		if _, ok := tags[list.Tags]; !ok {
			tags[list.Tags] = []item{}
		}
		tags[list.Tags] = append(tags[list.Tags], item{Name: list.As, IsPrefix: list.IsPrefix})
	}

	var keys tagSlice
	for key := range tags {
		if key == "" {
			continue
		} else {
			keys = append(keys, key)
		}
	}

	sort.Sort(keys)

	counter := 1
	for _, key := range keys {
		str += fmt.Sprintf(" *%s*\n", helpers.CapitalizeWords(key))
		for _, e := range tags[key] {
			var prefix string
			if e.IsPrefix {
				// Get prefix from the original message body
				prefix, _ = libs.ExtractPrefix(m.Body)
			} else {
				prefix = ""
			}
			for _, nm := range e.Name {
				str += fmt.Sprintf("%d. ```%s%s```\n", counter, prefix, nm)
				counter++
			}
		}
		str += "\n"
	}

	m.Reply(str)
	return true
}

func init() {
	libs.NewCommands(&libs.ICommand{
		Name:     "menu",
		As:       []string{"menu"},
		Tags:     "main",
		IsPrefix: true,
		Execute:  menu,
	})
}

