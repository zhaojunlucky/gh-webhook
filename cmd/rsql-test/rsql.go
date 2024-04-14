package main

import (
	log "github.com/sirupsen/logrus"
	"time"
)
import rsql "github.com/si3nloong/go-rsql"

type QueryParams struct {
	Name        string    `rsql:"n,filter,sort,allow=eq|gt|gte"`
	Status      string    `rsql:"status,filter,sort"`
	PtrStr      *string   `rsql:"text,filter,sort"`
	No          int       `rsql:"no,filter,sort,column=No2"`
	SubmittedAt time.Time `rsql:"submittedAt,filter"`
	CreatedAt   time.Time `rsql:"createdAt,sort"`
}

func main() {
	var i QueryParams
	p := rsql.MustNew(i)

	params, err := p.ParseQuery(`filter=status=eq="111";no=gt=1991;text==null&sort=status,-no`)
	if err != nil {
		panic(err)
	}

	log.Println(params.Filters)
	log.Println(params.Sorts)
}
