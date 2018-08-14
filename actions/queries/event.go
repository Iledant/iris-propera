package queries

// SQLGetAdminNextMonthEvents queries all events in the coming 30 days.
const SQLGetAdminNextMonthEvents = `SELECT e.id, e.date, o.name AS operation, e.name AS event FROM event e, physical_op o 
WHERE e.date < CURRENT_DATE + interval '1 month' AND e.date >= CURRENT_DATE 
	AND e.physical_op_id=o.id`

// SQLGetUserNextMonthEvents queries all events in the coming 30 days for operations belonging to the user.
const SQLGetUserNextMonthEvents = SQLGetAdminNextMonthEvents +
	` AND o.id IN (SELECT rights.physical_op_id FROM rights WHERE rights.users_id = ?)`
