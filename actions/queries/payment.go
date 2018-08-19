package queries

// GetPaymentsPerMonth calculate the payments per month of a given year.
const GetPaymentsPerMonth = `SELECT EXTRACT(YEAR FROM date) AS year, 
  EXTRACT(MONTH FROM date) AS month, SUM(value - cancelled_value) AS value
FROM payment WHERE date >= ? GROUP BY 1,2 ORDER BY 1,2`

// DeleteTempPayment clear the temporary table.
const DeleteTempPayment = `DELETE FROM temp_payment`

// InsertTempPayment insert a new value in the temporary table.
const InsertTempPayment = `INSERT INTO temp_payment VALUES (DEFAULT,?,?,?,?,?,?,?,?,?)`

// UpdatePaymentWithTemp update payment rows with new temporary values.
const UpdatePaymentWithTemp = `WITH new AS (
	SELECT p.id, t.number, t.date, t.value, t.cancelled_value FROM temp_payment t
		LEFT JOIN payment p ON t.number = p.number AND t.date = p.date
	WHERE (p.value <> t.value))
UPDATE payment SET value = new.value, cancelled_value = new.cancelled_value
FROM new WHERE payment.id = new.id`

// InsertTempIntoPayment insert temporary payments not already present.
const InsertTempIntoPayment = `INSERT INTO PAYMENT (financial_commitment_id, coriolis_year, coriolis_egt_code,
	coriolis_egt_num, coriolis_egt_line, date, number, value, cancelled_value, beneficiary_code)
	SELECT NULL financial_commitment_id, coriolis_year, coriolis_egt_code, coriolis_egt_num, 
 coriolis_egt_line, date, number, value, cancelled_value, beneficiary_code FROM temp_payment t
	WHERE (t.number, t.date) NOT IN (SELECT number, date FROM payment)`

// CalculatePaymentFcID calculate the link between payment and financial commitments.
const CalculatePaymentFcID = `WITH ref AS (
	SELECT DISTINCT ON (coriolis_year, coriolis_egt_code, coriolis_egt_num, coriolis_egt_line) 
	id, coriolis_year, coriolis_egt_code, coriolis_egt_num, coriolis_egt_line 
	FROM financial_commitment ORDER BY coriolis_year, coriolis_egt_code, coriolis_egt_num, coriolis_egt_line) 
	 UPDATE payment SET 
		 financial_commitment_id = ref.id 
	 FROM ref WHERE (payment.coriolis_year = ref.coriolis_year AND 
	payment.coriolis_egt_code = ref.coriolis_egt_code AND 
	payment.coriolis_egt_num = ref.coriolis_egt_num AND 
	payment.coriolis_egt_line = ref.coriolis_egt_line)`

// PrevisionRealized calculate the payment prevision and real payments for the given year and beneficiary.
const PrevisionRealized = `WITH pr AS (SELECT * FROM payment_ratios WHERE payment_types_id = ?),
fc_sum AS (SELECT beneficiary_code, EXTRACT(year FROM date)::integer AS year, SUM(value) AS value 
							FROM financial_commitment WHERE EXTRACT(year FROM date) < ? GROUP BY 1,2),
fc AS (SELECT fc_sum.beneficiary_code, SUM(fc_sum.value * pr.ratio) AS value 
					FROM fc_sum, pr WHERE fc_sum.year + pr.index = ? GROUP BY 1),
p AS (SELECT beneficiary_code, SUM(value) AS value 
					FROM payment WHERE EXTRACT(YEAR from date) = ? GROUP BY 1)
SELECT b.name, fc.value::bigint AS prev_payment, COALESCE(p.value,0) AS payment FROM fc
LEFT JOIN beneficiary b ON fc.beneficiary_code = b.code
LEFT OUTER JOIN p ON p.beneficiary_code = b.code
ORDER BY 2 DESC`

// monthCumulatedBegin is the common beggining of the query that calculates cumulated payment per month for one or all beneficiaries.
const monthCumulatedBegin = `SELECT tot.year, tot.month, sum(tot.value) OVER (PARTITION BY tot.year ORDER BY tot.month) as cumulated FROM
(SELECT extract(month from DATE) as month, EXTRACT (year FROM date) AS year, 0.01*sum(value) as value 
	 FROM payment `

// monthCumulatedEnd is the common end of the query that calculates cumulated payment per month for one or all beneficiaries.
const monthCumulatedEnd = `GROUP BY 1,2 ORDER BY 2,1) tot ORDER BY 1,2`

// MonthCumulatedAll calculates cumulated payment per month for all beneficiaries.
const MonthCumulatedAll = monthCumulatedBegin + monthCumulatedEnd

// MonthCumulatedBeneficiary calculates cumulated payment per month for a beneficiary.
const MonthCumulatedBeneficiary = monthCumulatedBegin + `WHERE beneficiary_code = ? ` + monthCumulatedEnd
