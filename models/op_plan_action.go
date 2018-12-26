package models

import (
	"database/sql"
	"strconv"
)

// OpWithPlanAndAction is used for the decoding the dedicated query.
type OpWithPlanAndAction struct {
	PhysicalOp
	PlanLineName       NullString `json:"plan_line_name"`
	PlanLineValue      NullInt64  `json:"plan_line_value"`
	PlanLineTotalValue NullInt64  `json:"plan_line_total_value"`
	BudgetActionName   NullString `json:"budget_action_name"`
	StepName           NullString `json:"step_name"`
	CategoryName       NullString `json:"category_name"`
}

// OpWithPlanAndActions embeddes an array of OpWithPlanAndAction for json export.
type OpWithPlanAndActions struct {
	OpWithPlanAndActions []OpWithPlanAndAction `json:"PhysicalOp"`
}

// GetAll fetches the operation with all informations linked according to role.
func (o *OpWithPlanAndActions) GetAll(uID int64, db *sql.DB) (err error) {
	from := "physical_op op"
	if uID != 0 {
		from = `(SELECT * FROM physical_op WHERE physical_op.id IN 
			(SELECT physical_op_id FROM rights WHERE users_id = ` + strconv.FormatInt(uID, 10) + `)) op `
	}
	rows, err := db.Query(`SELECT op.id, op.number, op.name, op.descript, op.isr, op.value,
		op.valuedate, op.length, op.tri, op.van, op.budget_action_id, op.payment_types_id, 
		op.plan_line_id, op.step_id, op.category_id, pl.name as plan_line_name, 
		pl.value as plan_line_value, pl.total_value as plan_line_total_value, 
		ba.name as budget_action_name, s.name AS step_name, c.name AS category_name 
		FROM ` + from + ` 
		LEFT OUTER JOIN budget_action ba ON op.budget_action_id = ba.id
		LEFT OUTER JOIN (SELECT pl.*, p.name AS plan_name FROM plan_line pl, plan p WHERE pl.plan_id = p.id) pl ON op.plan_line_id = pl.id
		LEFT OUTER JOIN plan p ON pl.plan_id = p.id
		LEFT OUTER JOIN step s ON op.step_id = s.id
		LEFT OUTER JOIN category c ON op.category_id = c.id`)
	if err != nil {
		return err
	}
	var r OpWithPlanAndAction
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&r.ID, &r.Number, &r.Name, &r.Descript, &r.Isr, &r.Value,
			&r.ValueDate, &r.Length, &r.TRI, &r.VAN, &r.BudgetActionID, &r.PaymentTypeID,
			&r.PlanLineID, &r.StepID, &r.CategoryID, &r.PlanLineName, &r.PlanLineValue,
			&r.PlanLineTotalValue, &r.BudgetActionName, &r.StepName, &r.CategoryName); err != nil {
			return err
		}
		o.OpWithPlanAndActions = append(o.OpWithPlanAndActions, r)
	}
	err = rows.Err()
	if len(o.OpWithPlanAndActions) == 0 {
		o.OpWithPlanAndActions = []OpWithPlanAndAction{}
	}
	return err
}

// Get fetches physical operation with fulls datas by ID from database.
func (op *OpWithPlanAndAction) Get(db *sql.DB) (err error) {
	err = db.QueryRow(`SELECT op.id, op.number, op.name, op.descript, op.isr, op.value,
	op.valuedate, op.length, op.tri, op.van, op.budget_action_id, op.payment_types_id, 
	op.plan_line_id, op.step_id, op.category_id, pl.name as plan_line_name, 
	pl.value as plan_line_value, pl.total_value as plan_line_total_value, 
	ba.name as budget_action_name, s.name AS step_name, c.name AS category_name 
	FROM physical_op op 
	LEFT OUTER JOIN budget_action ba ON op.budget_action_id = ba.id
	LEFT OUTER JOIN (SELECT pl.*, p.name AS plan_name FROM plan_line pl, plan p WHERE pl.plan_id = p.id) pl ON op.plan_line_id = pl.id
	LEFT OUTER JOIN plan p ON pl.plan_id = p.id
	LEFT OUTER JOIN step s ON op.step_id = s.id
	LEFT OUTER JOIN category c ON op.category_id = c.id WHERE op.id = $1`, op.ID).
		Scan(&op.ID, &op.Number, &op.Name, &op.Descript, &op.Isr, &op.Value,
			&op.ValueDate, &op.Length, &op.TRI, &op.VAN, &op.BudgetActionID, &op.PaymentTypeID,
			&op.PlanLineID, &op.StepID, &op.CategoryID, &op.PlanLineName, &op.PlanLineValue,
			&op.PlanLineTotalValue, &op.BudgetActionName, &op.StepName, &op.CategoryName)
	return err
}
