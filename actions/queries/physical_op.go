package queries

// Queries used in physical_op

//SQLGetOpWithPlanAction : query to get all physical operations with additional fields in plain text.
const SQLGetOpWithPlanAction = `SELECT op.*, pl.plan_name, pl.plan_id, pl.name as plan_line_name, pl.value as plan_line_value, 
pl.total_value as plan_line_total_value, ba.name as budget_action_name, s.name AS step_name,
c.name AS category_name 
FROM physical_op op
LEFT OUTER JOIN budget_action ba ON op.budget_action_id = ba.id
LEFT OUTER JOIN (SELECT pl.*, p.name AS plan_name FROM plan_line pl, plan p WHERE pl.plan_id = p.id) pl ON op.plan_line_id = pl.id
LEFT OUTER JOIN plan p ON pl.plan_id = p.id
LEFT OUTER JOIN step s ON op.step_id = s.id
LEFT OUTER JOIN category c ON op.category_id = c.id`

// SQLCreateTempPhysOp : query to create temporary physical op table for batch import.
const SQLCreateTempPhysOp = `CREATE TABLE temp_physical_op ( 
	number varchar(10) NOT NULL, name varchar(255) NOT NULL, descript text, 
	isr boolean, value bigint, valuedate date, length bigint, 
	tri integer, van bigint, action varchar(11), step varchar(50),
	category varchar(50), payment_types_id integer, plan_line_id integer)`

// SQLUpdateTempPhysOp : query to insert temporary physical operations into physical operations table.
const SQLUpdateTempPhysOp = `WITH new AS (
	SELECT p.id, t.number, t.name, t.descript, t.isr, t.value, t.valuedate, t.length, t.tri, t.van, 
				 b.id AS budget_action_id, t.payment_types_id, t.plan_line_id, s.id AS step_id, c.id AS category_id 
	FROM temp_physical_op t
	LEFT JOIN physical_op p ON t.number = p.number
	LEFT OUTER JOIN (SELECT ba.id,  (bp.code_contract || bp.code_function || bp.code_number || ba.code) AS code
		 FROM budget_action ba, budget_program bp 
		 WHERE ba.program_id = bp.id) b ON b.code = t.action
	LEFT OUTER JOIN step s ON s.name = t.step
	LEFT OUTER JOIN category c ON c.name = t.category)
UPDATE physical_op AS op SET 
	name = new.name, descript = COALESCE(new.descript, op.descript),  isr = COALESCE(new.isr, op.isr),
	value = COALESCE(new.value, op.value), valuedate = COALESCE(new.valuedate, op.valuedate),
	length = COALESCE(new.length, op.length), tri = COALESCE(new.tri, op.tri), van = COALESCE(new.van, op.van), 
	budget_action_id = COALESCE(new.budget_action_id, op.budget_action_id),
	payment_types_id = COALESCE(new.payment_types_id, op.payment_types_id),
	plan_line_id = COALESCE(new.plan_line_id, op.plan_line_id), step_id = COALESCE(new.step_id, op.step_id),
	category_id = COALESCE(new.category_id, op.category_id)
FROM new WHERE op.id = new.id;`

// SQLAddTempPhysOp : query to add temporary physical operations whose number not in physical operations table.
const SQLAddTempPhysOp = `INSERT INTO physical_op (number, name, descript, isr, value, valuedate, length,
	tri, van, payment_types_id, budget_action_id, plan_line_id, step_id, category_id)
SELECT t.number, t.name, t.descript, t.isr, t.value, t.valuedate, t.length, t.tri, t.van, t.payment_types_id, 
b.id AS budget_action_id, t.plan_line_id, s.id, c.id
FROM temp_physical_op t
LEFT OUTER JOIN (SELECT ba.id,  (bp.code_contract || bp.code_function || bp.code_number || ba.code) AS code
	FROM budget_action ba, budget_program bp
	WHERE ba.program_id = bp.id) b ON b.code = t.action
LEFT OUTER JOIN step s ON s.name = t.step
LEFT OUTER JOIN category c ON c.name = t.category
WHERE t.number NOT IN (SELECT number FROM physical_op);`
