package models

import (
	"database/sql"
	"fmt"
	"time"
)

// DifActionPmtPrevision model
type DifActionPmtPrevision struct {
	Chapter  int64   `json:"chapter"`
	ActionID int64   `json:"action_id"`
	Action   string  `json:"action"`
	Prev     float64 `json:"prev"`
	Y0       float64 `json:"y0"`
	Y1       float64 `json:"y1"`
	Y2       float64 `json:"y2"`
	Y3       float64 `json:"y3"`
	Y4       float64 `json:"y4"`
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
	Chapter  int64
	ActionID int64
	Action   string
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
	actionCmt as (SELECT extract(year FROM date) y,action_id,sum(value)::bigint v 
		FROM financial_commitment
		WHERE extract (year FROM date)>=2007
		AND extract(year FROM date)<extract(year FROM CURRENT_DATE)
	  	AND value > 0
		GROUP BY 1,2 order by 1,2),
	actionPmt as (SELECT extract(year FROM f.date) y,f.action_id,
		extract(year FROM p.date)-extract(year FROM f.date) as idx,sum(p.value) v
		FROM payment p
		JOIN financial_commitment f ON p.financial_commitment_id=f.id
		WHERE extract(year FROM f.date)>=2007
			AND extract(year FROM p.date)-extract(year FROM f.date)>=0
			AND extract(year FROM p.date)<extract(year FROM CURRENT_DATE)
		GROUP BY 1,2,3 order by 1,2,3),
	actionId as (select distinct action_id FROM actionCmt),
	y as (select generate_series(2007,extract(year from CURRENT_DATE)::int) y),
	idx as (select generate_series(0,max(idx)::int) idx from actionPmt idx),
	yidx as (select y.y,idx.idx from y,idx WHERE idx.idx+y.y<extract(year from current_date)),
	compActionPmt as (select yidx.y,actionID.action_id,yidx.idx,COALESCE(actionPmt.v,0) v
		FROM yidx
	  	CROSS JOIN actionID
		LEFT OUTER JOIN actionPmt ON actionPmt.y=yidx.y 
			AND actionPmt.idx=yidx.idx AND actionPmt.action_id=actionID.action_id
		order by 1,2,3
	),
	cumActionPmt as (SELECT y,action_id,idx,sum(v) 
		OVER (PARTITION by y,action_id ORDER BY y,action_id,idx) FROM compActionPmt),
	dry as (SELECT y,action_id,0 as idx,actionCmt.v::bigint FROM actionCmt
		UNION ALL
		SELECT cumActionPmt.y,actionCmt.action_id,cumActionPmt.idx+1,
			actionCmt.v-cumActionPmt.sum v 
		FROM actionCmt
		JOIN cumActionPmt on actionCmt.y=cumActionPmt.y AND 
			actionCmt.action_id=cumActionPmt.action_id
	),
	ramProg as (SELECT y,action_id,v FROM dry 
			WHERE y+idx=extract(year FROM CURRENT_DATE)
		UNION ALL
		SELECT p.year,op.budget_action_id action_id,sum(p.value) v
		FROM programmings p
		JOIN physical_op op on p.physical_op_id=op.id
		WHERE year=extract(year FROM CURRENT_DATE)
		GROUP BY 1,2
		UNION ALL
		SELECT year,action_id,v FROM
			(SELECT p.year,op.budget_action_id action_id,sum(p.value) v
				FROM prev_commitment p
				JOIN physical_op op on p.physical_op_id=op.id
				WHERE year>extract(year FROM CURRENT_DATE)
					AND year<extract(year FROM CURRENT_DATE)+5
				GROUP BY 1,2) prg 
		),
		years as (SELECT generate_series(min(y)::int,
			extract(year FROM CURRENT_DATE)::int+4) y FROM ramProg),
	aid as (SELECT distinct action_id FROM ramProg)
SELECT q.y,q.action_id,COALESCE(ramProg.v,0)::double precision*0.00000001
	FROM (SELECT years.y,aid.action_id FROM years,aid) q
	LEFT OUTER JOIN ramProg on q.y=ramProg.y AND q.action_id=ramProg.action_id
	WHERE q.action_id NOTNULL
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
	||COALESCE(bp.code_subfunction,'')||ba.code FROM budget_action ba
	JOIN budget_program bp ON ba.program_id=bp.id
	JOIN budget_chapter chap ON bp.chapter_id=chap.id
	ORDER BY 2`
	rows, err := db.Query(q)
	if err != nil {
		return fmt.Errorf("SELECT action datas %v", err)
	}
	var line actionItem
	for rows.Next() {
		if err = rows.Scan(&line.Chapter, &line.ActionID, &line.Action); err != nil {
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
		m.Lines[x].Action = actions.Lines[i].Action
	}
	return nil
}
