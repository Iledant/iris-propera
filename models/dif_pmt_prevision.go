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
	with 
  years as (SELECT * FROM generate_series(2007,EXTRACT(year FROM CURRENT_DATE)::int-1) y),
  fcy as (SELECT sum(value)::bigint as v, EXTRACT(year FROM date)::int as y
    FROM financial_commitment WHERE EXTRACT(year FROM date)<EXTRACT(year FROM CURRENT_DATE)
      AND EXTRACT (year FROM date)>=2007
		GROUP by 2 ORDER by 2),
  fcyn as (SELECT years.y,COALESCE(fcy.v,0::bigint) v
    FROM years LEFT OUTER JOIN fcy ON years.y=fcy.y),
  max_idx as (SELECT max(EXTRACT(year FROM p.date)-EXTRACT(year FROM f.date))::int as m
    FROM payment p
    JOIN financial_commitment f on p.financial_commitment_id=f.id
		WHERE EXTRACT(year FROM p.date)-EXTRACT(year FROM f.date)>=0
			AND EXTRACT(year FROM p.date)<EXTRACT(year FROM CURRENT_DATE)
      AND EXTRACT(year FROM f.date)>=2007),
	pmty as (SELECT sum(p.value)::bigint as v, EXTRACT(year FROM p.date)::int-
		EXTRACT(year FROM f.date)::int as idx,EXTRACT(year FROM f.date)::int as y
		FROM payment p
		JOIN financial_commitment f on p.financial_commitment_id=f.id
		WHERE EXTRACT(year FROM p.date)-EXTRACT(year FROM f.date)>=0
			AND EXTRACT(year FROM p.date)<EXTRACT(year FROM CURRENT_DATE)
      AND EXTRACT(year FROM f.date)>=2007
		GROUP by 2,3 ORDER by 3,2),
  idx as (select generate_series(0,m) i from max_idx),
  pmtyn as (SELECT years.y,idx.i,COALESCE(pmty.v,0) v from years
    CROSS join idx 
    LEFT JOIN pmty ON pmty.y=years.y AND pmty.idx=idx.i
    WHERE years.y+idx.i <= 2018
  order by 1,2),
	c_pmty as (SELECT y,i,sum(v) over (partition by y ORDER by y,i) FROM pmtyn),
  ram_y as (select sum(value)::bigint as v,EXTRACT(year FROM CURRENT_DATE)::int as y,0 as idx
		FROM programmings WHERE year=EXTRACT(year FROM CURRENT_DATE)
		UNION ALL
		SELECT fcyn.v-c_pmty.sum as v,fcyn.y as y,c_pmty.i+1 as idx FROM fcyn 
		JOIN c_pmty on fcyn.y=c_pmty.y),
  ratio_y as (SELECT ram_y.y,pmty.idx,pmty.v::double precision/ram_y.v as ratio
		FROM pmty,ram_y
		WHERE pmty.y=ram_y.y and ram_y.idx=pmty.idx and ram_y.y>=2008),
    avg_ratio as (select idx,avg(ratio) as ratio from ratio_y group by 1)
(SELECT COALESCE(SUM(ram_y.v*ratio_y.ratio)/100000000.0,0),ratio_y.y FROM ram_y, ratio_y
		WHERE ram_y.idx=ratio_y.idx AND ram_y.y+ram_y.idx=EXTRACT(year FROM CURRENT_DATE)
		GROUP by 2 ORDER by 2)
	UNION ALL
	(SELECT SUM(ram_y.v*avg_ratio.ratio)/100000000.0,0 FROM ram_y, avg_ratio
		WHERE ram_y.y+avg_ratio.idx=EXTRACT(year FROM CURRENT_DATE));`
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
	with fcy as (select extract (year from date) y,sum(value) v from financial_commitment
  where extract (year from date)>=2007
    and extract(year from date)<extract(year from current_date)
  group by 1 order by 1),
pmy as (select extract(year from f.date) y,
  extract(year from p.date)-extract(year from f.date) as idx, sum(p.value) v
  from payment p
  join financial_commitment f ON p.financial_commitment_id=f.id
  where extract(year from f.date)>=2007
    AND extract(year from p.date)-extract(year from f.date)>=0
    AND extract(year from p.date)<extract(year from CURRENT_DATE)
group by 1,2 order by 1,2),
spy as (select y,idx,sum(v) OVER (PARTITION by y ORDER BY y,idx) from pmy),
ry as (select y,0 as idx,fcy.v from fcy
UNION ALL
  select spy.y,spy.idx+1,fcy.v-spy.sum v from fcy join spy on fcy.y=spy.y
),
r as (select ry.y,ry.idx,COALESCE(pmy.v,0)::double precision/ry.v r
  FROM ry join pmy on ry.y=pmy.y and ry.idx=pmy.idx
  where ry.y<extract(year from current_date)
  )
select idx,avg(r) from r where idx+y>=extract(year from current_date) - 2
  group by 1 order by 1
`
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
	with fcy as (select extract (year from date) y,sum(value) v from financial_commitment
  where extract (year from date)>=2007
    and extract(year from date)<extract(year from current_date)
  group by 1 order by 1),
pmy as (select extract(year from f.date) y,
  extract(year from p.date)-extract(year from f.date) as idx, sum(p.value) v
  from payment p
  join financial_commitment f ON p.financial_commitment_id=f.id
  where extract(year from f.date)>=2007
    AND extract(year from p.date)-extract(year from f.date)>=0
    AND extract(year from p.date)<extract(year from CURRENT_DATE)
group by 1,2 order by 1,2),
spy as (select y,idx,sum(v) OVER (PARTITION by y ORDER BY y,idx) from pmy),
ry as (select y,0 as idx,fcy.v from fcy
UNION ALL
  select spy.y,spy.idx+1,fcy.v-spy.sum v from fcy join spy on fcy.y=spy.y
),
r as (select ry.y,ry.idx,COALESCE(pmy.v,0)::double precision/ry.v r
  FROM ry join pmy on ry.y=pmy.y and ry.idx=pmy.idx
  where ry.y<extract(year from current_date)
  ),
avr as (select idx,avg(r) from r where idx+y>=extract(year from current_date) - 2
  group by 1 order by 1)
	select y,v*0.00000001::double precision from
	(select * from ry
  UNION ALL
  select year,0 idx,sum(value) from programmings 
  where year=extract(year from current_date)
	group by 1,2) q
  where y+idx = extract(year from current_date)
	order by 1
`
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

	actualYear := time.Now().Year()

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
	for i < 5-len(prev) {
		ram = append(ram, yearVal{Year: int64(y + i), Val: 0})
		i++
	}
	j = ramLen - 1
	for y = 0; y < 5; y++ {
		p.Year = int64(y + actualYear)
		p.Prev = 0
		for i = 0; i < ratioLen && j-i > 0; i++ {
			q := ratios[i] * ram[j-i].Val
			p.Prev += q
			ram[j-i].Val -= q
		}
		j++
		m.Lines = append(m.Lines, p)
	}
	return nil
}
