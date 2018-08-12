package queries

// SQLDropTempActionTable drop if exists the temporary table.
const SQLDropTempActionTable = `DROP TABLE IF EXISTS temp_actions`

// SQLCreateTempActionTable create temporary table to import on array of budget actions.
const SQLCreateTempActionTable = `CREATE TABLE temp_actions ( code_contract VARCHAR(1), 
	code_function VARCHAR(2), code_number VARCHAR(3),  action_code VARCHAR(4), 
	name VARCHAR(255),sector VARCHAR(10));`

// SQLInsertTempAction is used to insert individual values in temporary budget actions table.
const SQLInsertTempAction = `INSERT INTO temp_actions (code_contract, code_function, 
		code_number, action_code, name, sector) VALUES (?, ?, ?, ?, ?, ?)`

// SQLUpdateBudgetAction update the name with temporary table of an action whose code already exists.
const SQLUpdateBudgetAction = `WITH new AS (
	SELECT a.id, t.name FROM temp_actions t, budget_program p, budget_action a
	WHERE t.action_code = a.code AND t.code_contract = p.code_contract AND
				t.code_function = p.code_function AND t.code_number = p.code_number AND
				a.program_id = p.id)
UPDATE budget_action SET
	name = new.name
FROM new WHERE budget_action.id = new.id`

// SQLInsertBudgetAction insert all actions in temporary table whose code doesn't already exists.
const SQLInsertBudgetAction = `INSERT INTO budget_action (program_id, sector_id, code, name) 
SELECT p.id AS program_id, s.id AS sector_id, t.action_code, t.name FROM temp_actions t
	LEFT JOIN budget_sector s ON s.code = t.sector
	LEFT JOIN budget_program p ON ( p.code_contract = t.code_contract AND
																	p.code_function = t.code_function AND
																	p.code_number = t.code_number)
WHERE (s.id, p.id, t.action_code) NOT IN (SELECT sector_id, program_id, code FROM budget_action) 
	AND p.id NOTNULL`
