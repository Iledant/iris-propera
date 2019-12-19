package models

import (
	"database/sql"
	"fmt"
	"time"
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

// MultiannualDifPmtPrevision model
type MultiannualDifPmtPrevision struct {
	Year int64   `json:"year"`
	Prev float64 `json:"prev"`
}

// MultiannualDifPmtPrevisions embeddes an array of MultiannualDifPmtPrevision
// for json export and calculation
type MultiannualDifPmtPrevisions struct {
	Lines []MultiannualDifPmtPrevision `json:"MultiannualDifPmtPrevision"`
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

func getDifRatios(db *sql.DB) ([]float64, error) {
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
		WHERE pmt_y.y=ram_y.y and ram_y.idx=pmt_y.idx and ram_y.y>=2008)
	select idx,avg(ratio) as ratio from ratio_y group by 1 order by 1`
	rows, err := db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("select ratio %v", err)
	}
	var (
		idx    int64
		ratio  float64
		ratios []float64
	)
	for rows.Next() {
		if err = rows.Scan(&idx, &ratio); err != nil {
			return nil, fmt.Errorf("scan ratio %v", err)
		}
		ratios = append(ratios, ratio)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err ratio %v", err)
	}
	return ratios, nil
}

type yearVal struct {
	Year int64
	Val  float64
}

func getRAM(db *sql.DB) ([]yearVal, error) {
	q := `
WITH fcy AS (SELECT EXTRACT(year FROM date) y, SUM(value) v
  FROM financial_commitment 
  WHERE EXTRACT(year from date)<EXTRACT(year FROM CURRENT_DATE)
  GROUP by 1),
pmy AS (SELECT EXTRACT(year FROM fc.date) y,SUM(p.value) v
  FROM payment p
  JOIN financial_commitment fc ON p.financial_commitment_id=fc.id
  WHERE EXTRACT(year from fc.date)<EXTRACT(year FROM CURRENT_DATE)
  GROUP BY 1)
SELECT fcy.y,((fcy.v-COALESCE(pmy.v,0))*0.00000001)::double precision
	FROM fcy LEFT JOIN pmy ON fcy.y=pmy.y ORDER BY 1;`
	rows, err := db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("select ram %v", err)
	}
	var (
		r  yearVal
		rr []yearVal
	)
	for rows.Next() {
		if err = rows.Scan(&r.Year, &r.Val); err != nil {
			return nil, fmt.Errorf("scan ram %v", err)
		}
		rr = append(rr, r)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err ram %v", err)
	}
	return rr, nil
}

func getProg(db *sql.DB) (float64, error) {
	q := `
		SELECT COALESCE(SUM(value),0)::double precision*0.00000001 FROM programmings
		WHERE year=EXTRACT(year FROM current_date)`
	var p float64
	if err := db.QueryRow(q).Scan(&p); err != nil {
		return 0, fmt.Errorf("query prog %v", err)
	}
	return p, nil
}

func getPrev(db *sql.DB) ([]yearVal, error) {
	q := `
	SELECT year,SUM(value)::double precision*0.00000001 FROM prev_commitment 
	WHERE year>EXTRACT(year FROM CURRENT_DATE)
	GROUP BY 1 ORDER BY 1;`
	rows, err := db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("select prev %v", err)
	}
	var (
		r  yearVal
		rr []yearVal
	)
	for rows.Next() {
		if err = rows.Scan(&r.Year, &r.Val); err != nil {
			return nil, fmt.Errorf("scan prev %v", err)
		}
		rr = append(rr, r)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err prev %v", err)
	}
	return rr, nil
}

// Get calculates the MultiannualDifPmtPrevision using the average differential
// ratios
func (m *MultiannualDifPmtPrevisions) Get(db *sql.DB) error {
	ratios, err := getDifRatios(db)
	if err != nil {
		return err
	}
	ratioLen := len(ratios)
	ram, err := getRAM(db)
	if err != nil {
		return err
	}
	ramLen := len(ram)

	prog, err := getProg(db)
	if err != nil {
		return err
	}
	actualYear := time.Now().Year()
	ram = append(ram, yearVal{Year: int64(actualYear), Val: prog})

	prev, err := getPrev(db)
	if err != nil {
		return err
	}
	for _, p := range prev {
		ram = append(ram, p)
	}
	var (
		p       MultiannualDifPmtPrevision
		i, j, y int
	)
	y = actualYear + len(prev) + 1
	fmt.Printf("Ratios : \n  %+v\nRam :\n  %+v\n", ratios, ram)
	for i < 5-len(prev) {
		ram = append(ram, yearVal{Year: int64(y + i), Val: 0})
		i++
	}
	j = ramLen
	for y = 0; y < 5; y++ {
		p.Year = int64(y + actualYear)
		p.Prev = 0
		for i = 0; i < ratioLen; i++ {
			q := ratios[i] * ram[j-i].Val
			p.Prev += q
			ram[j-i].Val -= q
		}
		j++
		m.Lines = append(m.Lines, p)
	}
	return nil
}
