package models

import (
	"fmt"
	"runtime"
	"testing"

	dbModels "github.com/HomelessHunter/CTC/db/models"
)

func printAlloc() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("%d KB\n", m.Alloc/1024)
}

func TestMapAllocation(t *testing.T) {
	session := NewSession()
	printAlloc()

	for i := 0; i < 1_000_000; i++ {
		alert, _ := dbModels.NewAlert(dbModels.WithPair(fmt.Sprint(i)), dbModels.WithMarket("binance"), dbModels.WithTargetPrice(58000.23))

		session.AddAlerts(1, *alert)
	}
	printAlloc()

	// fmt.Println(len(*session.alerts[1]))
	alertT, _ := dbModels.NewAlert(dbModels.WithPair(fmt.Sprint(100)), dbModels.WithMarket("binance"))
	alertT1, _ := dbModels.NewAlert(dbModels.WithPair(fmt.Sprint(876)), dbModels.WithMarket("binance"))
	alerts := *session.alerts[1]
	i, err := alertT.Find(alerts)
	if err != nil {
		t.Error(err)
	}
	session.DeleteAlert(1, i)
	in, err := alertT1.Find(alerts)
	if err != nil {
		t.Error(err)
	}
	session.DeleteAlert(1, in)

	session.DeleteAlerts(1, len(alerts))

	runtime.GC()
	printAlloc()
	runtime.KeepAlive(session)
}

type Foo struct {
	v []byte
}

func TestSliceAllocation(t *testing.T) {
	foos := make([]Foo, 1_000)
	printAlloc()

	for i := 0; i < len(foos); i++ {
		foos[i] = Foo{
			v: make([]byte, 1024*1024),
		}
	}
	printAlloc()

	foos = keepFirstTwoElementsOnly(foos)
	runtime.GC()
	printAlloc()
	runtime.KeepAlive(foos)
}

func keepFirstTwoElementsOnly(foos []Foo) []Foo {
	res := make([]Foo, 2)
	copy(res, foos)
	return res
}
