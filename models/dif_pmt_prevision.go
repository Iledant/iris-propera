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

// DifPmtPrevisions embeddes an array of DifPmtPrevision
// for json export and calculation
type DifPmtPrevisions struct {
	Lines []DifPmtPrevision `json:"DifPmtPrevision"`
}

func getDifRatios(db *sql.DB) ([]float64, error) {
	q := `
	with fcy as (select extract (year from date) y,sum(value)::bigint v
	from financial_commitment
	where extract (year from date)>=2007
		and extract(year from date)<extract(year from current_date)
	and value>0
	group by 1 order by 1),
	pmy as (select extract(year from f.date) y,
		extract(year from p.date)-extract(year from f.date) as idx, sum(p.value) v
	from payment p
	join financial_commitment f ON p.financial_commitment_id=f.id
	where extract(year from f.date)>=2007
		AND extract(year from p.date)-extract(year from f.date)>=0
		AND extract(year from p.date)<extract(year from CURRENT_DATE)
group by 1,2),
y as (select generate_series(2007,extract(year from CURRENT_DATE)::int) y),
idx as (select generate_series(0,max(idx)::int) idx from pmy idx),
cpmy as (select y.y,idx.idx,COALESCE(v,0)::bigint v from y
	cross join idx
	left outer join pmy on pmy.y=y.y and idx.idx=pmy.idx
	where y.y+idx.idx<extract(year from current_date)
	order by 1,2),
spy as (select y,idx,sum(v) OVER (PARTITION by y ORDER BY y,idx) from cpmy),
ry as (select y,0 as idx,fcy.v from fcy
UNION ALL
	select spy.y,spy.idx+1,fcy.v-spy.sum v from fcy join spy on fcy.y=spy.y
),
r as (select ry.y,ry.idx,COALESCE(cpmy.v,0)::double precision/ry.v r
	FROM ry join cpmy on ry.y=cpmy.y and ry.idx=cpmy.idx
	where ry.y<extract(year from current_date)
	)
select idx,avg(r) from r where idx+y>=extract(year from current_date) - 2
	group by 1 order by 1`
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
	with fcy as (select extract (year from date) y,sum(value)::bigint v
	from financial_commitment
	where extract (year from date)>=2007
		and extract(year from date)<extract(year from current_date)
	and value>0
	group by 1 order by 1),
	pmy as (select extract(year from f.date) y,
		extract(year from p.date)-extract(year from f.date) as idx, sum(p.value) v
	from payment p
	join financial_commitment f ON p.financial_commitment_id=f.id
	where extract(year from f.date)>=2007
		AND extract(year from p.date)-extract(year from f.date)>=0
		AND extract(year from p.date)<extract(year from CURRENT_DATE)
group by 1,2),
y as (select generate_series(2007,extract(year from CURRENT_DATE)::int) y),
idx as (select generate_series(0,max(idx)::int) idx from pmy idx),
cpmy as (select y.y,idx.idx,COALESCE(v,0)::bigint v from y
	cross join idx
	left outer join pmy on pmy.y=y.y and idx.idx=pmy.idx
	where y.y+idx.idx<extract(year from current_date)
	order by 1,2),
spy as (select y,idx,sum(v) OVER (PARTITION by y ORDER BY y,idx) from cpmy),
ry as (select y,0 as idx,fcy.v from fcy
UNION ALL
	select spy.y,spy.idx+1,fcy.v-spy.sum v from fcy join spy on fcy.y=spy.y
)
select y,v*0.00000001::double precision from
	(select * from ry
	UNION ALL
	select year,0 idx,sum(value) from programmings
	where year=extract(year from current_date)
	group by 1,2) q
	where y+idx = extract(year from current_date)
	order by 1`

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

// Get calculates the DifPmtPrevision using the average differential
// ratios
func (m *DifPmtPrevisions) Get(db *sql.DB) error {
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
		p       DifPmtPrevision
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
		for i = 0; i < ratioLen && j-i >= 0; i++ {
			q := ratios[i] * ram[j-i].Val
			p.Prev += q
			ram[j-i].Val -= q
		}
		j++
		m.Lines = append(m.Lines, p)
	}
	return nil
}
