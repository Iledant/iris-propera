package queries

// getPreProgPrefix is the common first part of the query to get the pre programmation.
const getPreProgPrefix = `SELECT op.id AS physical_op_id, op.number AS physical_op_number, op.name AS physical_op_name,
pc.value AS prev_value, pc.state_ratio AS prev_state_ratio, 
pc.total_value AS prev_total_value, pc.descript AS prev_descript, pp.id AS pre_prog_id,
pp.value AS pre_prog_value, pp.year AS pre_prog_year, pp.commission_id AS pre_prog_commission_id, 
pp.state_ratio AS pre_prog_state_ratio, pp.total_value AS pre_prog_total_value, 
pp.descript AS pre_prog_descript, pl.plan_name, pl.plan_line_name, pl.plan_line_value, pl.plan_line_total_value 
FROM`

// getPreProgFromAdminClause is the middle part of the query to get the pre programmation i.e. for admin all physical operations.
const getPreProgFromAdminClause = ` physical_op op `

// getPreProgFromUserClause is the middle part of the query to get the pre programmation i.e. par user his physical operations.
const getPreProgFromUserClause = ` (SELECT * FROM physical_op WHERE id IN (SELECT physical_op_id FROM rights WHERE users_id = ?)) op `

// getPreProgSuffix is the common last part of the query to get the pre programmation.
const getPreProgSuffix = `LEFT OUTER JOIN (SELECT pl.id, pl.name AS plan_line_name, pl.value AS plan_line_value, pl.total_value AS plan_line_total_value, p.name AS plan_name
					FROM plan_line pl, plan p WHERE pl.plan_id = p.id) pl ON op.plan_line_id = pl.id
LEFT OUTER JOIN (SELECT * FROM prev_commitment WHERE year = ?) pc ON op.id = pc.physical_op_id
LEFT OUTER JOIN (SELECT * FROM pre_programmings WHERE year = ?) pp ON op.id = pp.physical_op_id`

// GetAdminPreProg is the query to get the pre programmation for all physical operations.
const GetAdminPreProg = getPreProgPrefix + getPreProgFromAdminClause + getPreProgSuffix

// GetUserPreProg is the query to get the pre programmation for one user's physical operations.
const GetUserPreProg = getPreProgPrefix + getPreProgFromUserClause + getPreProgSuffix

// CreateTempPreProgTable creates the temporary table to import pre programming.
const CreateTempPreProgTable = `CREATE TABLE IF NOT EXISTS temp_pre_programmings 
(id integer, year integer NOT NULL, physical_op_id integer NOT NULL, commission_id integer NOT NULL,
	value bigint NOT NULL, total_value bigint, state_ratio double precision, descript text)`

// InsertTempPreProg insert sent data into temporary table for pre programming.
const InsertTempPreProg = `INSERT INTO temp_pre_programmings 
(id, year, physical_op_id, commission_id, value, total_value, state_ratio, descript) 
VALUES (?, ?, ?, ?, ?, ?, ?, NULL)`

// UpdatePreProgWithTemp updates pre programming value whose id is in temp table.
const UpdatePreProgWithTemp = `UPDATE pre_programmings SET
	 value = t.value, physical_op_id = t.physical_op_id, commission_id = t.commission_id,
	 year = t.year, total_value = t.total_value, state_ratio = t.state_ratio, descript = t.descript
FROM temp_pre_programmings t WHERE pre_programmings.id = t.id`

// delPreProgAdminPart select all physical operations.
const delPreProgAdminPart = `(SELECT id FROM physical_op op)`

// delPreProgUserPart select only physical operations belonging to the user.
const delPreProgUserPart = `(SELECT id FROM physical_op
	WHERE id IN (SELECT physical_op_id FROM rights WHERE users_id = ?))`

// DelPreProgAdmin delete former pre programming not present in temp data for an admin.
const DelPreProgAdmin = `DELETE FROM pre_programmings pp WHERE pp.physical_op_id IN ` + delPreProgAdminPart +
	` AND pp.id NOT IN (SELECT id FROM temp_pre_programmings) AND pp.year = ?`

// DelPreProgUser delete former pre programming not present in temp data for a user.
const DelPreProgUser = `DELETE FROM pre_programmings pp WHERE pp.physical_op_id IN ` + delPreProgUserPart +
	` AND pp.id NOT IN (SELECT id FROM temp_pre_programmings) AND pp.year = ?`

// InsertPreProg insert new pre programming from temporary table to permanent one.
const InsertPreProg = `INSERT INTO pre_programmings (value, physical_op_id, commission_id, year, 
	total_value, state_ratio, descript)
(SELECT value, physical_op_id, commission_id, year, total_value, state_ratio, descript 
	FROM temp_pre_programmings WHERE id NOT IN (SELECT DISTINCT id FROM pre_programmings))`

// DeleteTempPreProgTable delete the temporary table used to import data.
const DeleteTempPreProgTable = `DROP TABLE IF EXISTS temp_pre_programmings`
