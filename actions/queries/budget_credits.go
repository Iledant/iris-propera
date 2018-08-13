package queries

// SQLGetMostRecentCredits gets the most recent budget credits according to commission date.
const SQLGetMostRecentCredits = `SELECT * FROM budget_credits WHERE commission_date = 
(SELECT max(commission_date) FROM budget_credits WHERE EXTRACT (year FROM commission_date) = ?)`

// SQLDropTempCreditsTable drop if exists temporary table for credits batch imports.
const SQLDropTempCreditsTable = `DROP TABLE IF EXISTS temp_budget_credits`

// SQLCreateTempCreditsTable create temporary table for credits batch imports.
const SQLCreateTempCreditsTable = `CREATE TABLE temp_budget_credits 
	(commission_date date, chapter integer, primary_commitment bigint, 
	frozen_commitment bigint, reserved_commitment bigint)`

// SQLInsertTempCredits inserts batch into credits temporary table.
const SQLInsertTempCredits = `INSERT INTO temp_budget_credits 
(commission_date, chapter, primary_commitment, reserved_commitment, frozen_commitment) 
VALUES (?, ?, ?, ?, ?)`

// SQLUpdateBatchCredits inserts only new batch credits into database.
const SQLUpdateBatchCredits = `INSERT INTO budget_credits
(commission_date, chapter_id, primary_commitment, frozen_commitment, reserved_commitment)
SELECT t.commission_date, bc.id, t.primary_commitment, t.frozen_commitment, t.reserved_commitment
FROM temp_budget_credits t
LEFT JOIN budget_chapter bc ON t.chapter = bc.code
WHERE (t.commission_date, t.chapter) NOT IN
(SELECT b.commission_date, c.code FROM budget_credits b, budget_chapter c WHERE b.chapter_id = c.id)`
