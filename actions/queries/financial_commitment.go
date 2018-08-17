package queries

// getUlFcs common header to get unlinked financial commitments.
const getUlFcs = `SELECT f.id as id, f.value as value, f.iris_code as iris_code, 
f.name as name, f.date as date, b.name as beneficiary 
FROM financial_commitment f, beneficiary b
WHERE f.beneficiary_code = b.code AND`

const paginateFoot = `ORDER BY 1 LIMIT 15 OFFSET ?`

// opSearch is part of the WHERE clause with search pattern for physical operations.
const opSearch = ` f.date >= ? AND physical_op_id ISNULL AND
(f.name ILIKE ? OR b.name ILIKE ? OR f.iris_code ILIKE ?)`

// opSearch is part of the WHERE clause with search pattern for plan lines.
const plSearch = ` f.date >= ? AND plan_line_id ISNULL AND
(f.name ILIKE ? OR b.name ILIKE ? OR f.iris_code ILIKE ?)`

// countUlFcs common header to count unlinked financial commitments.
const countUlFcs = `SELECT count(f.id) count FROM financial_commitment f, beneficiary b WHERE f.beneficiary_code = b.code AND`

// SQLCountOpUnlinkedFcs gets the number of financial commitments not linked to physical operation and that match search pattern.
const SQLCountOpUnlinkedFcs = countUlFcs + opSearch

// SQLCountPlUnlinkedFcs gets the number of financial commitments not linked to plan line and that match search pattern.
const SQLCountPlUnlinkedFcs = countUlFcs + plSearch

// SQLGetOpUnlinkedFcs is used to get physical operation unlinked financial commitments.
const SQLGetOpUnlinkedFcs = getUlFcs + opSearch + paginateFoot

// SQLGetPlUnlinkedFcs is used to get plan line unlinked financial commitments.
const SQLGetPlUnlinkedFcs = getUlFcs + plSearch + paginateFoot

// SQLCountOpLinkedFcs gets the number of financial commitments linked to a physical operation and matching the search pattern
const SQLCountOpLinkedFcs = `SELECT count(f.id) 
FROM financial_commitment f, beneficiary b, physical_op op
WHERE f.physical_op_id = op.id AND f.beneficiary_code = b.code AND f.physical_op_id NOTNULL
AND f.date > ? AND (f.name ILIKE ? OR b.name ILIKE ? OR op.name ILIKE ? 
OR op.number ILIKE ?)`

// SQLGetOpLinkedFcs is used to get physical operation linked financial commitments.
const SQLGetOpLinkedFcs = `SELECT f.id as fc_iD, f.value as fc_value, f.name as fc_name, 
f.iris_code, f.date as fc_date, b.Name fc_beneficiary, op.number op_number, op.name op_name
FROM financial_commitment f, beneficiary b, physical_op op
WHERE f.physical_op_id = op.id AND f.beneficiary_code = b.code AND f.physical_op_id NOTNULL
AND f.date > ? AND (f.name ILIKE ? OR b.name ILIKE ? OR op.name ILIKE ? 
OR op.number ILIKE ?)` + paginateFoot

// SQLCountPlLinkedFcs is used to get physical operation linked financial commitments.
const SQLCountPlLinkedFcs = `SELECT count(f.id) 
FROM financial_commitment f, beneficiary b, plan_line pl
WHERE f.plan_line_id = pl.id AND f.beneficiary_code = b.code AND f.plan_line_id NOTNULL
AND f.date > ? AND (f.name ILIKE ? OR b.name ILIKE ? OR pl.name ILIKE ?)`

// SQLGetPlLinkedFcs is used to get physical operation linked financial commitments.
const SQLGetPlLinkedFcs = `SELECT f.id as fc_id, f.value as fc_value, f.name as fc_name, 
f.iris_code, f.date as fc_date, b.Name fc_beneficiary, pl.name pl_name
FROM financial_commitment f, beneficiary b, plan_line pl
WHERE f.plan_line_id = pl.id AND f.beneficiary_code = b.code AND f.plan_line_id NOTNULL
AND f.date > ? AND (f.name ILIKE ? OR b.name ILIKE ? OR pl.name ILIKE ?)` + paginateFoot

// SQLMonthFCs gets the amount of financial commitment of each month for the year given in parameter.
const SQLMonthFCs = `SELECT extract(month from date) AS month, sum(value) AS value
FROM financial_commitment WHERE extract(year FROM date) = ? GROUP BY 1 ORDER BY 1`

// SQLUpdateUploadFcs updates the financial commitments table with uploaded datas into temp_commitment.
const SQLUpdateUploadFcs = `WITH new AS (
	SELECT f.id, t.chapter, t.action, t.iris_code, t.name, t.beneficiary_code, t.date, t.value, t.lapse_date
	FROM temp_commitment t LEFT JOIN financial_commitment f ON t.iris_code = f.iris_code 
	 WHERE (f.value <> t.value OR f.chapter <> t.chapter OR f.action <> t.action OR f.name <> t.name OR 
					 f.coriolis_year <> t.coriolis_year OR  f.coriolis_egt_code <> t.coriolis_egt_code OR 
					 f.coriolis_egt_num <> t.coriolis_egt_num OR f.coriolis_egt_line <> t.coriolis_egt_line OR 
					 f.beneficiary_code <> t.beneficiary_code OR f.lapse_date IS DISTINCT FROM t.lapse_date) 
					 AND f.date = t.date) 
UPDATE financial_commitment SET 
chapter = new.chapter,  action = new.action, name = new.name, beneficiary_code = new.beneficiary_code, 
 value = new.value, lapse_date = new.lapse_date 
FROM new WHERE financial_commitment.id = new.id`

// SQLInsertUploadFcs inserts into the financial commitments table uploaded datas not present in that table.
const SQLInsertUploadFcs = `INSERT INTO financial_commitment (physical_op_id, chapter, action, iris_code,
	coriolis_year, coriolis_egt_code, coriolis_egt_num, coriolis_egt_line, name, beneficiary_code, date,
	value, lapse_date) 
SELECT NULL as physical_op_id, chapter, action, iris_code, coriolis_year, coriolis_egt_code, coriolis_egt_num, 
	coriolis_egt_line, name, beneficiary_code, date, value, lapse_date
	FROM temp_commitment t 
WHERE (t.iris_code, t.date) NOT IN (SELECT iris_code, date FROM financial_commitment)`

// SQLInsertNewBeneficiary insert uploaded beneficiaries into beneficaries table if not present.
const SQLInsertNewBeneficiary = `WITH new AS (
	SELECT t.beneficiary_code, t.beneficiary, t.date FROM temp_commitment t
	WHERE t.beneficiary_code NOT IN (SELECT code FROM beneficiary) )
INSERT INTO beneficiary (code, name) SELECT beneficiary_code, beneficiary FROM new
  WHERE (date, beneficiary_code) IN (SELECT Max(date), beneficiary_code FROM temp_commitment GROUP BY 2)`

// SQLZeroDuplicatedFcs set to 0 financial commitments value when they appears after a change of beneficiary.
const SQLZeroDuplicatedFcs = ` WITH duplicated AS (SELECT id from financial_commitment WHERE iris_code IN
	(SELECT iris_code FROM financial_commitment WHERE iris_code in
		(SELECT iris_code FROM
			(SELECT SUM(1) as count, iris_code FROM financial_commitment GROUP BY 2) fcCount WHERE fcCount.count > 1)
					AND coriolis_egt_line <> '1') AND coriolis_egt_line = '1')
UPDATE financial_commitment SET value = 0 FROM duplicated WHERE financial_commitment.id=duplicated.id`

// SQLUpdateFcActionField calculates the action id field of uploaded financial commitments.²
const SQLUpdateFcActionField = `WITH correspond AS (SELECT fc_extract.fc_id, ba_full.ba_id FROM 
	(SELECT fc.id AS fc_id, substring (fc.action FROM '^[0-9sS]+') AS fc_action FROM financial_commitment fc) fc_extract,
(SELECT ba.id AS ba_id, bp.code_contract || bp.code_function || bp.code_number || ba.code AS ba_code 
FROM budget_action ba, budget_program bp WHERE ba.program_id = bp.id) ba_full
WHERE fc_extract.fc_action = ba_full.ba_code)
UPDATE financial_commitment SET action_id = correspond.ba_id
FROM correspond WHERE financial_commitment.id = correspond.fc_id`

// SQLBatchOpFc update the physical_op_id field of uploaded financial commitments.
const SQLBatchOpFc = `UPDATE financial_commitment SET physical_op_id = op.id
FROM physical_op op, temp_attachment WHERE op.number = temp_attachment.op_number AND 
financial_commitment.coriolis_year=temp_attachment.coriolis_year AND 
financial_commitment.coriolis_egt_code =temp_attachment.coriolis_egt_code AND
financial_commitment.coriolis_egt_num=temp_attachment.coriolis_egt_num AND
financial_commitment.coriolis_egt_line=temp_attachment.coriolis_egt_line`

// SQLInsertTempCommitment insert into temp commitment uploaded item.
const SQLInsertTempCommitment = `INSERT INTO temp_commitment (chapter,action, 
	iris_code,coriolis_year,coriolis_egt_code,coriolis_egt_num,coriolis_egt_line,name,
	beneficiary,beneficiary_code,date,value,lapse_date) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`
