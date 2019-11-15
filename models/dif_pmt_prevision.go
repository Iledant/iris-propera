package models

import (
	"database/sql"
	"fmt"
)

// DifPmtPrevision model
type DifPmtPrevision struct {
	Year int64   `json:"year"`
	Prev float64 `json:"prev"`
}

// DifPmtPrevisions embeddes an array of DifPmtPrevision for json export and query
type DifPmtPrevisions struct {
	Lines []DifPmtPrevision `json:"DifPmtPrevision"`
}

// Get calculates the paiement prevision of the current year
// using differential ratios (i.e. applied to the difference between commitments
// and payments) for the commitments of different years and for the avarage of
// this ratios
func (d *DifPmtPrevisions) Get(db *sql.DB) error {
	q := `
	with fcy as ((SELECT sum(value)::bigint as v, EXTRACT(year FROM date)::int as y
		FROM financial_commitment where extract(year from date)<EXTRACT(year FROM CURRENT_DATE)
		GROUP by 2 ORDER by 2)
		UNION ALL
		(select sum(value)::bigint as v,EXTRACT(year FROM CURRENT_DATE)::int as y
		FROM programmings WHERE year=EXTRACT(year FROM CURRENT_DATE))    ),
	pmt_y as (SELECT sum(p.value)::bigint as v, EXTRACT(year FROM p.date)::int-
		EXTRACT(year FROM f.date)::int as idx,EXTRACT(year FROM f.date)::int as y
		FROM payment p JOIN financial_commitment f on p.financial_commitment_id=f.id
		WHERE EXTRACT(year FROM p.date)-EXTRACT(year FROM f.date)>=0
		GROUP by 2,3 ORDER by 3,2),
	c_pmt_y as (SELECT y,idx,sum(v) over (partition by y ORDER by y,idx) FROM pmt_y),
	ram_y as (SELECT fcy.v as v,fcy.y as y,0 as idx FROM fcy
		UNION ALL
		SELECT fcy.v-c_pmt_y.sum as v,fcy.y as y,c_pmt_y.idx+1 as idx FROM fcy 
		JOIN c_pmt_y on fcy.y=c_pmt_y.y),
	ratio_y as (SELECT ram_y.y,pmt_y.idx,pmt_y.v::double precision/ram_y.v as ratio
		FROM pmt_y,ram_y
		WHERE pmt_y.y=ram_y.y and ram_y.idx=pmt_y.idx and ram_y.y>=2008),
	avg_ratio as (select idx,avg(ratio) as ratio from ratio_y group by 1)
	(SELECT COALESCE(SUM(ram_y.v*ratio_y.ratio)/100000000.0,0),ratio_y.y FROM ram_y, ratio_y
		WHERE ram_y.idx=ratio_y.idx AND ram_y.y+ram_y.idx=EXTRACT(year FROM CURRENT_DATE)
		GROUP by 2 ORDER by 2)
	UNION ALL
	(SELECT SUM(ram_y.v*avg_ratio.ratio)/100000000.0,0 FROM ram_y, avg_ratio
		WHERE ram_y.idx=avg_ratio.idx and ram_y.y+avg_ratio.idx=EXTRACT(year FROM CURRENT_DATE));`
	rows, err := db.Query(q)
	if err != nil {
		return fmt.Errorf("select %v", err)
	}
	var l DifPmtPrevision
	for rows.Next() {
		if err = rows.Scan(&l.Prev, &l.Year); err != nil {
			return fmt.Errorf("scan %v", err)
		}
		d.Lines = append(d.Lines, l)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("rows err %v", err)
	}
	if len(d.Lines) == 0 {
		d.Lines = []DifPmtPrevision{}
	}
	return nil
}
