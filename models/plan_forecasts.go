package models

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

// PlanForecast model
type PlanForecast struct {
	Number    string
	Name      string
	TRI       NullInt64
	VAN       NullInt64
	Value     NullInt64
	ValueDate NullInt64
	Step      NullString
	Category  NullString
	R75       NullFloat64
	R77       NullFloat64
	R78       NullFloat64
	R91       NullFloat64
	R92       NullFloat64
	R93       NullFloat64
	R94       NullFloat64
	R95       NullFloat64
	TotalPrev []int64
	Prev      []int64
}

// PlanForecasts embeddes an array of PlanForecast for json export and dedicated
// queries
type PlanForecasts struct {
	Lines []PlanForecast `json:"PlanForecast"`
}

// GetAll fetches physical operations caracteristics, department ratios and
// commitment previsions between two years
func (p *PlanForecasts) GetAll(db *sql.DB, firstYear, lastYear int64) error {
	if lastYear < firstYear {
		return fmt.Errorf("lastYear inférieure à firstYear")
	}
	i := firstYear
	var array1, array2, years []string
	for i <= lastYear {
		array1 = append(array1, fmt.Sprintf("COALESCE(q1.y%d,0)", i))
		array2 = append(array2, fmt.Sprintf("COALESCE(q2.y%d,0)", i))
		years = append(years, fmt.Sprintf("y%d BIGINT", i))
		i++
	}
	query := fmt.Sprintf(`SELECT op.number,op.name,op.tri,op.van,op.value,op.valuedate,
		step.name,category.name,odr.r75,odr.r77,odr.r78,odr.r91,odr.r92,odr.r93,
		odr.r94,odr.r95,q.total_value,q.value
	FROM physical_op op
	LEFT JOIN step ON op.step_id=step.id
	LEFT JOIN category ON op.category_id=category.id
	JOIN
		(SELECT q1.op_id,ARRAY[%s] as value,ARRAY[%s] as total_value
		FROM
			(SELECT * FROM crosstab('SELECT physical_op_id,year,value
				FROM prev_commitment WHERE year>=%d AND year<=%d')
			AS ct(op_id integer,%s))q1
		JOIN
			(SELECT * FROM crosstab('SELECT physical_op_id,year,total_value
				FROM prev_commitment WHERE year>=%d AND year<=%d')
			AS ct(op_id integer,%s))q2
		ON q1.op_id=q2.op_id) q ON q.op_id=op.id
	LEFT JOIN op_dpt_ratios odr ON op.id=odr.physical_op_id`,
		strings.Join(array1, ","), strings.Join(array2, ","), firstYear, lastYear,
		strings.Join(years, ","), firstYear, lastYear, strings.Join(years, ","))
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("query %v", err)
	}
	var line PlanForecast
	for rows.Next() {
		if err := rows.Scan(&line.Number, &line.Name, &line.TRI, &line.VAN,
			&line.Value, &line.ValueDate, &line.Step, &line.Category, &line.R75,
			&line.R77, &line.R78, &line.R91, &line.R92, &line.R93, &line.R94,
			&line.R95, pq.Array(&line.TotalPrev), pq.Array(&line.Prev)); err != nil {
			return fmt.Errorf("scan %v", err)
		}
		p.Lines = append(p.Lines, line)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("rows err %v", err)
	}
	if len(p.Lines) == 0 {
		p.Lines = []PlanForecast{}
	}
	return nil
}
