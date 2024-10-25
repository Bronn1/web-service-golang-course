package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

const filePath2 string = "./data/users.txt"

type User struct {
	Browsers []string `json:"browsers"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
}

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		panic("cannot read file")
	}
	foundUsers := ""
	uniqueBrowsers := 0
	seenBrowsers := map[string]bool{}
	for i, line := range strings.Split(string(content), "\n") {
		user := &User{}
		if err := json.Unmarshal([]byte(line), user); err != nil {
			fmt.Println("wrong json format: ", err)
			continue
		}
		isAndroid := false
		isMSIE := false
		for _, browser := range user.Browsers {
			if strings.Contains(browser, "Android") {
				isAndroid = true
				if _, ok := seenBrowsers[browser]; !ok {
					seenBrowsers[browser] = true
					uniqueBrowsers++
				}
			}
			if strings.Contains(browser, "MSIE") {
				isMSIE = true
				if _, ok := seenBrowsers[browser]; !ok {
					seenBrowsers[browser] = true
					uniqueBrowsers++
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, strings.Replace(user.Email, "@", " [at] ", 1))
	}

	fmt.Fprintln(out, "found users:\n"+foundUsers)
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
	//fmt.Println(content)
}
