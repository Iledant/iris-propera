package models

import (
	"database/sql"
	"fmt"
)

// PmtPrevision model
type PmtPrevision struct {
	Year int64   `json:"year"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
}

// PmtPrevisions embeddes an array of PmtPrevision for json export and query
type PmtPrevisions struct {
	Lines []PmtPrevision `json:"PmtPrevision"`
}

// Get calculates the paiement prevision of the current year
// using direct ratios (i.e. applied to the difference between commitments
// and payments) for the commitments of different years and for the avarage of
// this ratios
func (d *PmtPrevisions) Get(db *sql.DB) error {
	q := `
	with fcy as ((select sum(value)::bigint as v, extract(year from date)::int as y
		FROM financial_commitment where extract(year from date)<EXTRACT(year FROM CURRENT_DATE)
		group by 2 order by 2)
    ),
	pmt_y as (select sum(p.value)::bigint as v, extract(year from p.date)::int-
			extract(year from f.date)::int as idx,extract(year from f.date)::int as y
		from payment p join financial_commitment f on p.financial_commitment_id=f.id
		where extract(year from p.date)-extract(year from f.date)>=0
      and extract(year from p.date)< extract(year from current_date)
		group by 2,3 order by 3,2),
	ratio_y as (select fcy.y,pmt_y.idx,pmt_y.v::double precision/fcy.v as ratio
	from pmt_y,fcy
	where pmt_y.y=fcy.y and fcy.y>=2007 and fcy.y<= extract(year from current_date)-5),
  apy as (select y,v from fcy
  UNION ALL
  select extract(year FROM current_date)::int y,sum(value)::bigint v FROM programmings 
    WHERE year=extract(year from current_date) group by 1
  UNION ALL
  select year y,sum(value)::bigint v from prev_commitment 
    where year>extract(year from current_date)group by 1),
  yp as (select y from apy where y>=extract(year FROM current_date) 
    and y <=extract(year from current_date)+4)
	select y,min(p),max(p) from
  (select yp.y,ratio_y.y as ry,sum(ratio_y.ratio*apy.v)::double precision/100000000.0 p
	from apy, ratio_y, yp where apy.y+ratio_y.idx = yp.y
	group by 1,2) q
  group by 1 order by 1
`
	rows, err := db.Query(q)
	if err != nil {
		return fmt.Errorf("select %v", err)
	}
	var l PmtPrevision
	for rows.Next() {
		if err = rows.Scan(&l.Year, &l.Min, &l.Max); err != nil {
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
