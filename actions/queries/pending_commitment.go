package queries

// DeletePendingTempTable drop the temporary table used to import batch pendings.
const DeletePendingTempTable = `DROP TABLE IF EXISTS temp_pending`

// CreatePendingTempTable creates the temporary table used to import batch pendings.
const CreatePendingTempTable = `CREATE TABLE temp_pending (
	chapter VARCHAR(5), action VARCHAR(154), iris_code VARCHAR(32),
	name VARCHAR(200), beneficiary VARCHAR(200), commission_date DATE,
	proposed_value BIGINT)`

// InsertBatchPending insert sent datas into temporary pending commitments.
const InsertBatchPending = `INSERT INTO temp_pending (chapter, action,
	iris_code, name, beneficiary, commission_date,proposed_value)
	VALUES (?,?,?,?,?,?,?)`

// UpdatePendingWithBatch updates the pendings commitments with imported batch values.
const UpdatePendingWithBatch = `UPDATE pending_commitments 
SET chapter = tp.chapter, action = tp.action, name = tp.name,
		beneficiary = tp.beneficiary, commission_date = tp.commission_date,
		proposed_value = tp.proposed_value
FROM (SELECT * FROM temp_pending) tp WHERE tp.iris_code = pending_commitments.iris_code`

// InsertPendingWithBatch inserts into the pending commitments table the imported batch values with a new iris_code.
const InsertPendingWithBatch = `INSERT INTO pending_commitments 
(physical_op_id, chapter, action, iris_code, name,  beneficiary, commission_date, proposed_value)
SELECT NULL,* FROM temp_pending WHERE iris_code NOT IN (SELECT iris_code FROM pending_commitments)`

// DeletePendingOutOfBatch deletes from pending commitments table rows when iris_code not in the imported batch.
const DeletePendingOutOfBatch = `DELETE FROM pending_commitments 
WHERE iris_code NOT IN (SELECT iris_code FROM temp_pending)`
