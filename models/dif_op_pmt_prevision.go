package models

import (
	"database/sql"
	"fmt"
	"time"
)

// DifOpPmtPrevision model
type DifOpPmtPrevision struct {
	OpID      int64     `json:"op_id"`
	OpNumber  string    `json:"op_number"`
	OpName    string    `json:"op_name"`
	OpChapter NullInt64 `json:"op_chapter"`
	Prev      float64   `json:"prev"`
	Y0        float64   `json:"y0"`
	Y1        float64   `json:"y1"`
	Y2        float64   `json:"y2"`
	Y3        float64   `json:"y3"`
	Y4        float64   `json:"y4"`
}

// DifOpPmtPrevisions embeddes an array of DifOpPmtPrevision for json
// export and dedicated queries
type DifOpPmtPrevisions struct {
	Lines []DifOpPmtPrevision `json:"DifOpPmtPrevision"`
}

type yearOpVal struct {
	Year int64
	OpID int64
	Val  float64
}

type opItem struct {
	Chapter  NullInt64
	OpID     int64
	OpNumber string
	OpName   string
}

type opItems struct {
	Lines []opItem
}

// getOpRAM computes the queries with all years and action IDs including
// the 4 comming years. The query using outer and cross joins to generate
// all value or zero in order for the algorithm to work without any further
// test
func getOpRAM(db *sql.DB) ([]yearOpVal, error) {
	q := `
	WITH
		cmt AS (SELECT extract(year FROM date) y,physical_op_id,sum(value)::bigint v 
			FROM financial_commitment
			WHERE extract (year FROM date)>=2007
			AND extract(year FROM date)<extract(year FROM CURRENT_DATE)
				AND value > 0
			GROUP BY 1,2),
		pmt AS (SELECT extract(year FROM f.date) y,f.physical_op_id,sum(p.value) v
			FROM payment p
			JOIN financial_commitment f ON p.financial_commitment_id=f.id
			WHERE extract(year FROM f.date)>=2007
				AND extract(year FROM p.date)-extract(year FROM f.date)>=0
				AND extract(year FROM p.date)<extract(year FROM CURRENT_DATE)
			GROUP BY 1,2),
		prg AS (SELECT p.year y,physical_op_id,sum(p.value)::bigint v
			FROM programmings p
			WHERE year=extract(year FROM CURRENT_DATE)
			GROUP BY 1,2),
		prev AS (SELECT year y,physical_op_id,v FROM
			(SELECT p.year,p.physical_op_id,sum(p.value)::bigint v
					FROM prev_commitment p
					WHERE year>extract(year FROM CURRENT_DATE)
						AND year<extract(year FROM CURRENT_DATE)+5
					GROUP BY 1,2) q),
		ram AS (SELECT cmt.y,cmt.physical_op_id,(cmt.v-COALESCE(pmt.v,0)::bigint) v FROM cmt
			LEFT OUTER JOIN pmt ON cmt.y=pmt.y AND cmt.physical_op_id=pmt.physical_op_id
			UNION ALL
			SELECT y,physical_op_id,v FROM prg
			UNION ALL
			SELECT y,physical_op_id,v FROM prev
		),
		op_id AS (SELECT distinct physical_op_id FROM ram),
		years AS (SELECT generate_series(2007,
			extract(year FROM current_date)::int+4)::int y)
	SELECT years.y,op_id.physical_op_id,COALESCE(ram.v,0)::double precision*0.00000001
	FROM op_id
	CROSS JOIN years
	LEFT OUTER JOIN ram ON ram.physical_op_id=op_id.physical_op_id AND ram.y=years.y
	WHERE op_id.physical_op_id NOTNULL
	ORDER BY 1,2`
	rows, err := db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("SELECT op ram %v", err)
	}
	var (
		r  yearOpVal
		rr []yearOpVal
	)
	for rows.Next() {
		if err = rows.Scan(&r.Year, &r.OpID, &r.Val); err != nil {
			return nil, fmt.Errorf("scan op ram %v", err)
		}
		rr = append(rr, r)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err op ram %v", err)
	}
	return rr, nil
}

func (o *opItems) Get(db *sql.DB) error {
	q := `SELECT q.code,op.id,op.number,op.name FROM physical_op op
	LEFT JOIN 
  (SELECT ba.id,chap.code 
    FROM budget_action ba 
	  JOIN budget_program bp ON ba.program_id=bp.id
	  JOIN budget_chapter chap ON bp.chapter_id=chap.id) q
    ON q.id=op.budget_action_id
	ORDER BY 2`
	rows, err := db.Query(q)
	if err != nil {
		return fmt.Errorf("SELECT op datas %v", err)
	}
	var line opItem
	for rows.Next() {
		if err = rows.Scan(&line.Chapter, &line.OpID, &line.OpNumber,
			&line.OpName); err != nil {
			return fmt.Errorf("scan op ram %v", err)
		}
		o.Lines = append(o.Lines, line)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("rows err op ram %v", err)
	}
	return nil
}

// Get calculates the DifOpPmtPrevision using the average differential
// ratios
func (m *DifOpPmtPrevisions) Get(db *sql.DB) error {
	ratios, err := getDifRatios(db)
	if err != nil {
		return err
	}
	ratioLen := len(ratios)
	ram, err := getOpRAM(db)
	if err != nil {
		return err
	}
	actualYear := time.Now().Year()
	var (
		opLen, actualYearBegin, j int
		p                         DifOpPmtPrevision
	)
	for i := 1; i < len(ram); i++ {
		if ram[i].OpID < ram[i-1].OpID {
			opLen = i
			break
		}
	}
	if opLen == 0 {
		return fmt.Errorf("impossible de trouver la séquence d'opérations dans la requête")
	}
	for i, a := range ram {
		if a.Year == int64(actualYear) {
			actualYearBegin = i
			break
		}
	}
	if actualYearBegin == 0 {
		return fmt.Errorf("impossible de trouver l'année en cours dans la requête")
	}
	var prev float64
	m.Lines = make([]DifOpPmtPrevision, opLen, opLen)
	for y := 0; y < 5; y++ {
		for a := 0; a < opLen; a++ {
			prev = 0
			j = actualYearBegin + a + y*opLen
			p.OpID = ram[j].OpID
			for i := 0; i < ratioLen && j-i*opLen >= 0; i++ {
				q := ratios[i] * ram[j-i*opLen].Val
				prev += q
				ram[j-i*opLen].Val -= q
			}
			m.Lines[a].OpID = ram[j].OpID
			switch y {
			case 0:
				m.Lines[a].Y0 = prev
			case 1:
				m.Lines[a].Y1 = prev
			case 2:
				m.Lines[a].Y2 = prev
			case 3:
				m.Lines[a].Y3 = prev
			case 4:
				m.Lines[a].Y4 = prev
			}
		}
	}
	var ops opItems
	if err = ops.Get(db); err != nil {
		return err
	}
	var i int
	opLen = len(ops.Lines)
	for x := 0; x < len(m.Lines); x++ {
		i = 0
		j = opLen - 1
		for {
			if m.Lines[x].OpID == ops.Lines[i].OpID {
				break
			}
			if m.Lines[x].OpID == ops.Lines[j].OpID {
				i = j
				break
			}
			if m.Lines[x].OpID < ops.Lines[(i+j)/2].OpID {
				j = (i + j) / 2
			} else {
				i = (i + j) / 2
			}
		}
		m.Lines[x].OpChapter = ops.Lines[i].Chapter
		m.Lines[x].OpNumber = ops.Lines[i].OpNumber
		m.Lines[x].OpName = ops.Lines[i].OpName
	}
	return nil
}
