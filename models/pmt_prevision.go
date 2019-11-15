package models

import (
	"database/sql"
	"fmt"
)

// PmtPrevision model
type PmtPrevision struct {
	Year int64   `json:"year"`
	Prev float64 `json:"prev"`
}

// PmtPrevisions embeddes an array of PmtPrevision for json export and query
type PmtPrevisions struct {
	Lines []PmtPrevision `json:"PmtPrevision"`
}

// Get calculates the paiement prevision of the current year
// using differential ratios (i.e. applied to the difference between commitments
// and payments) for the commitments of different years and for the avarage of
// this ratios
func (d *PmtPrevisions) Get(db *sql.DB) error {
	q := `
	with fcy as ((select sum(value)::bigint as v, extract(year from date)::int as y
		FROM financial_commitment where extract(year from date)<EXTRACT(year FROM CURRENT_DATE)
		group by 2 order by 2)
		UNION ALL
		(select sum(value)::bigint as v,EXTRACT(year FROM CURRENT_DATE)::int as y
		FROM programmings WHERE year=EXTRACT(year FROM CURRENT_DATE))),
	pmt_y as (select sum(p.value)::bigint as v, extract(year from p.date)::int-
			extract(year from f.date)::int as idx,extract(year from f.date)::int as y
		from payment p join financial_commitment f on p.financial_commitment_id=f.id
		where extract(year from p.date)-extract(year from f.date)>=0
		group by 2,3 order by 3,2),
	ratio_y as (select fcy.y,pmt_y.idx,pmt_y.v::double precision/fcy.v as ratio
	from pmt_y,fcy
	where pmt_y.y=fcy.y and fcy.y>=2008),
	avg_ratio as (select idx,avg(ratio) as ratio from ratio_y group by 1)
	(select ratio_y.y,sum(ratio_y.ratio*fcy.v)::double precision/100000000.0
	from fcy, ratio_y where fcy.y+ratio_y.idx = EXTRACT(year FROM CURRENT_DATE)
	group by 1 order by 1)
	union all
	(SELECT 0,sum(avg_ratio.ratio*fcy.v)::double precision/100000000.0
	from fcy, avg_ratio where fcy.y+avg_ratio.idx = EXTRACT(year FROM CURRENT_DATE));`
	rows, err := db.Query(q)
	if err != nil {
		return fmt.Errorf("select %v", err)
	}
	var l PmtPrevision
	for rows.Next() {
		if err = rows.Scan(&l.Year, &l.Prev); err != nil {
			return fmt.Errorf("scan %v", err)
		}
		d.Lines = append(d.Lines, l)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("rows err %v", err)
	}
	if len(d.Lines) == 0 {
		d.Lines = []PmtPrevision{}
	}
	return nil
}
