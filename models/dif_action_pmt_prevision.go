package models

import (
	"database/sql"
	"fmt"
	"time"
)

// DifActionPmtPrevision model
type DifActionPmtPrevision struct {
	Chapter    int64   `json:"chapter"`
	ActionID   int64   `json:"action_id"`
	ActionCode string  `json:"action_code"`
	ActionName string  `json:"action_name"`
	Prev       float64 `json:"prev"`
	Y0         float64 `json:"y0"`
	Y1         float64 `json:"y1"`
	Y2         float64 `json:"y2"`
	Y3         float64 `json:"y3"`
	Y4         float64 `json:"y4"`
}

// DifActionPmtPrevisions embeddes an array of DifActionPmtPrevision for json
// export and dedicated queries
type DifActionPmtPrevisions struct {
	Lines []DifActionPmtPrevision `json:"DifActionPmtPrevision"`
}

type yearActionVal struct {
	Year     int64
	ActionID int64
	Val      float64
}

type actionItem struct {
	Chapter    int64
	ActionID   int64
	ActionCode string
	ActionName string
}

type actionItems struct {
	Lines []actionItem
}

// getActionRAM computes the queries with all years and action IDs including
// the 4 comming years. The query using outer and cross joins to generate
// all value or zero in order for the algorithm to work without any further
// test
func getActionRAM(db *sql.DB) ([]yearActionVal, error) {
	q := `
	WITH
		cmt AS (SELECT extract(year FROM date) y,action_id,sum(value)::bigint v 
			FROM financial_commitment
			WHERE extract (year FROM date)>=2007
			AND extract(year FROM date)<extract(year FROM CURRENT_DATE)
				AND value > 0
			GROUP BY 1,2),
		pmt AS (SELECT extract(year FROM f.date) y,f.action_id,sum(p.value) v
			FROM payment p
			JOIN financial_commitment f ON p.financial_commitment_id=f.id
			WHERE extract(year FROM f.date)>=2007
				AND extract(year FROM p.date)-extract(year FROM f.date)>=0
				AND extract(year FROM p.date)<extract(year FROM CURRENT_DATE)
			GROUP BY 1,2),
		prg AS (SELECT p.year y,op.budget_action_id action_id,sum(p.value)::bigint v
			FROM programmings p
			JOIN physical_op op on p.physical_op_id=op.id
			WHERE year=extract(year FROM CURRENT_DATE)
			GROUP BY 1,2),
		prev AS (SELECT year y,action_id,v FROM
			(SELECT p.year,op.budget_action_id action_id,sum(p.value)::bigint v
					FROM prev_commitment p
					JOIN physical_op op on p.physical_op_id=op.id
					WHERE year>extract(year FROM CURRENT_DATE)
						AND year<extract(year FROM CURRENT_DATE)+5
					GROUP BY 1,2) q),
		ram AS (SELECT cmt.y,cmt.action_id,(cmt.v-COALESCE(pmt.v,0)::bigint) v FROM cmt
			LEFT OUTER JOIN pmt ON cmt.y=pmt.y AND cmt.action_id=pmt.action_id
			UNION ALL
			SELECT y,action_id,v FROM prg
			UNION ALL
			SELECT y,action_id,v FROM prev
		),
		action_id AS (SELECT distinct action_id FROM ram),
		years AS (SELECT generate_series(2007,
			extract(year FROM current_date)::int+4)::int y)
	SELECT years.y,action_id.action_id,COALESCE(ram.v,0)::double precision*0.00000001
	FROM action_id
	CROSS JOIN years
	LEFT OUTER JOIN ram ON ram.action_id=action_id.action_id AND ram.y=years.y
	WHERE action_id.action_id NOTNULL
	ORDER BY 1,2`
	rows, err := db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("SELECT action ram %v", err)
	}
	var (
		r  yearActionVal
		rr []yearActionVal
	)
	for rows.Next() {
		if err = rows.Scan(&r.Year, &r.ActionID, &r.Val); err != nil {
			return nil, fmt.Errorf("scan action ram %v", err)
		}
		rr = append(rr, r)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err action ram %v", err)
	}
	return rr, nil
}

func (a *actionItems) Get(db *sql.DB) error {
	q := `SELECT chap.code,ba.id,bp.code_contract||bp.code_function||bp.code_number
	||COALESCE(bp.code_subfunction,'')||ba.code,ba.name FROM budget_action ba
	JOIN budget_program bp ON ba.program_id=bp.id
	JOIN budget_chapter chap ON bp.chapter_id=chap.id
	ORDER BY 2`
	rows, err := db.Query(q)
	if err != nil {
		return fmt.Errorf("SELECT action datas %v", err)
	}
	var line actionItem
	for rows.Next() {
		if err = rows.Scan(&line.Chapter, &line.ActionID, &line.ActionCode,
			&line.ActionName); err != nil {
			return fmt.Errorf("scan action ram %v", err)
		}
		a.Lines = append(a.Lines, line)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("rows err action ram %v", err)
	}
	return nil
}

// Get calculates the DifActionPmtPrevision using the average differential
// ratios
func (m *DifActionPmtPrevisions) Get(db *sql.DB) error {
	ratios, err := getDifRatios(db)
	if err != nil {
		return err
	}
	ratioLen := len(ratios)
	ram, err := getActionRAM(db)
	if err != nil {
		return err
	}
	actualYear := time.Now().Year()
	var (
		actionLen, actualYearBegin, j int
		p                             DifActionPmtPrevision
	)
	for i := 1; i < len(ram); i++ {
		if ram[i].ActionID < ram[i-1].ActionID {
			actionLen = i
			break
		}
	}
	if actionLen == 0 {
		return fmt.Errorf("impossible de trouver la séquence d'action dans la requête")
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
	m.Lines = make([]DifActionPmtPrevision, actionLen, actionLen)
	for y := 0; y < 5; y++ {
		for a := 0; a < actionLen; a++ {
			prev = 0
			j = actualYearBegin + a + y*actionLen
			p.ActionID = ram[j].ActionID
			for i := 0; i < ratioLen && j-i*actionLen >= 0; i++ {
				q := ratios[i] * ram[j-i*actionLen].Val
				if i+int(ram[j-i*actionLen].Year) != actualYear+y {
					fmt.Printf("différence de ratio+year : %+v Année : %d\n", ram[j-i*actionLen], actualYear+y)
				}
				prev += q
				ram[j-i*actionLen].Val -= q
			}
			m.Lines[a].ActionID = ram[j].ActionID
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
	var actions actionItems
	if err = actions.Get(db); err != nil {
		return err
	}
	var i int
	actionLen = len(actions.Lines)
	for x := 0; x < len(m.Lines); x++ {
		i = 0
		j = actionLen - 1
		for {
			if m.Lines[x].ActionID == actions.Lines[i].ActionID {
				break
			}
			if m.Lines[x].ActionID == actions.Lines[j].ActionID {
				i = j
				break
			}
			if m.Lines[x].ActionID < actions.Lines[(i+j)/2].ActionID {
				j = (i + j) / 2
			} else {
				i = (i + j) / 2
			}
		}
		m.Lines[x].Chapter = actions.Lines[i].Chapter
		m.Lines[x].ActionCode = actions.Lines[i].ActionCode
		m.Lines[x].ActionName = actions.Lines[i].ActionName
	}
	return nil
}
