package common

import (
	"strings"
	"sync"
)

func WaitGoroutines(wg *sync.WaitGroup, usersChanges chan UserProjects) {
	wg.Wait()
	close(usersChanges)
}

func TrimEmail(email string) string {
	return strings.Split(email, "@")[0]
}
