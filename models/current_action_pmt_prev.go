package models

import (
	"database/sql"
	"fmt"
)

// CurYearActionPmtPrevision model
type CurYearActionPmtPrevision struct {
	ActionID     sql.NullInt64
	Chapter      NullString
	Sector       NullString
	Function     NullString
	ActionCode   NullString
	ActionName   NullString
	PmtPrevision float64
	Payment      float64
}

// CurYearActionPmtPrevisions embeddes an array of CurYearActionPmtPrevision for
// json export AND dedicated queries
type CurYearActionPmtPrevisions struct {
	Lines []CurYearActionPmtPrevision `json:"CurYearActionPmtPrevision"`
}

// Get fetches the payment AND differential ratios method payment prevision for
// the current year
func (c *CurYearActionPmtPrevisions) Get(db *sql.DB) error {
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
			JOIN physical_op op ON p.physical_op_id=op.id
			WHERE year=extract(year FROM CURRENT_DATE)
			GROUP BY 1,2),
		prev AS (SELECT year y,action_id,v FROM
			(SELECT p.year,op.budget_action_id action_id,sum(p.value)::bigint v
					FROM prev_commitment p
					JOIN physical_op op ON p.physical_op_id=op.id
					WHERE year>extract(year FROM CURRENT_DATE)
						AND year<extract(year FROM CURRENT_DATE)+5
					GROUP BY 1,2) q),
		ram AS (SELECT cmt.y,cmt.action_id,(cmt.v-COALESCE(pmt.v,0)::bigint) v FROM cmt
			LEFT OUTER JOIN pmt ON cmt.y=pmt.y AND cmt.action_id=pmt.action_id
			UNION ALL
			SELECT y,action_id,v FROM prg
		),
		action_id AS (SELECT distinct action_id FROM ram),
		years AS (SELECT generate_series(2007,
			extract(year FROM current_date)::int)::int y),
		full_ram as (SELECT years.y,action_id.action_id,
				COALESCE(ram.v,0)::double precision*0.00000001 v
			FROM action_id
			CROSS JOIN years
			LEFT OUTER JOIN ram ON ram.action_id=action_id.action_id AND ram.y=years.y
			WHERE action_id.action_id NOTNULL),
		fcy as (SELECT extract (year FROM date) y,sum(value)::bigint v
		FROM financial_commitment
		WHERE extract (year FROM date)>=2007
			AND extract(year FROM date)<extract(year FROM current_date)
			AND value>0
		GROUP BY 1),
		pmy as (SELECT extract(year FROM f.date) y,
			extract(year FROM p.date)-extract(year FROM f.date) as idx, sum(p.value) v
			FROM payment p
			JOIN financial_commitment f ON p.financial_commitment_id=f.id
			WHERE extract(year FROM f.date)>=2007
				AND extract(year FROM p.date)-extract(year FROM f.date)>=0
				AND extract(year FROM p.date)<extract(year FROM CURRENT_DATE)
			GROUP BY 1,2),
		idx as (SELECT generate_series(0,max(idx)::int) idx FROM pmy idx),
		cpmy as (SELECT years.y,idx.idx,COALESCE(v,0)::bigint v FROM years
			CROSS JOIN idx
			LEFT OUTER JOIN pmy ON pmy.y=years.y AND idx.idx=pmy.idx
			WHERE years.y+idx.idx<extract(year FROM current_date)
			ORDER BY 1,2),
		spy as (SELECT y,idx,sum(v) OVER (PARTITION BY y ORDER BY y,idx) FROM cpmy),
		ry as (SELECT y,0 as idx,fcy.v FROM fcy
			UNION ALL
			SELECT spy.y,spy.idx+1,fcy.v-spy.sum v FROM fcy JOIN spy ON fcy.y=spy.y),
		r as (SELECT ry.y,ry.idx,COALESCE(cpmy.v,0)::double precision/ry.v r
			FROM ry JOIN cpmy ON ry.y=cpmy.y AND ry.idx=cpmy.idx
			WHERE ry.y<extract(year FROM current_date)),
		avg_ratio as (SELECT idx,avg(r) FROM r 
			WHERE idx+y>=extract(year FROM current_date) - 2
			GROUP BY 1),
		actual_pmt as (SELECT f.action_id,sum(p.value)::double precision*0.00000001 v
			FROM payment p 
			LEFT OUTER JOIN financial_commitment f ON p.financial_commitment_id=f.id
			WHERE extract(year FROM p.date)=extract(year FROM current_date)
			GROUP BY 1)
	SELECT action_id.action_id,chap.code,bs.code,bp.code_function||
		COALESCE(bp.code_subfunction,''),bp.code_contract||bp.code_function||
		bp.code_number||COALESCE(bp.code_subfunction,''),ba.name,
		COALESCE(q.v,0)::double precision,COALESCE(actual_pmt.v,0)
	FROM action_id
	JOIN budget_action ba ON action_id.action_id=ba.id
	JOIN budget_program bp ON ba.program_id=bp.id
	JOIN budget_chapter chap ON bp.chapter_id=chap.id
	JOIN budget_sector bs ON ba.sector_id=bs.id
	LEFT OUTER JOIN (SELECT f.action_id,sum(f.v::double precision*a.avg) v
					FROM full_ram f JOIN avg_ratio a ON f.y+a.idx=extract(year FROM CURRENT_DATE)
				GROUP BY 1) q
	ON q.action_id=action_id.action_id
	FULL OUTER JOIN actual_pmt ON action_id.action_id=actual_pmt.action_id
	ORDER BY 1;`
	rows, err := db.Query(q)
	if err != nil {
		return fmt.Errorf("SELECT ratio %v", err)
	}
	var line CurYearActionPmtPrevision
	for rows.Next() {
		if err = rows.Scan(&line.ActionID, &line.Chapter, &line.Sector,
			&line.Function, &line.ActionCode, &line.ActionName, &line.PmtPrevision,
			&line.Payment); err != nil {
			return fmt.Errorf("scan ratio %v", err)
		}
		c.Lines = append(c.Lines, line)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("rows err ratio %v", err)
	}
	return nil
}
