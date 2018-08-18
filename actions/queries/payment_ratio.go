package queries

// SQLDeleteRatios delete all ratios linked to a payment type.
const SQLDeleteRatios = "DELETE from payment_ratios WHERE payment_types_id = ?"

// SQLGetYearRatio calculate the ratios of payments for the financial commitments of a given year.
const SQLGetYearRatio = `WITH yc AS (SELECT f.id FROM financial_commitment f WHERE f.coriolis_year = ?),
total AS (SELECT sum(f.value) as total FROM financial_commitment f WHERE f.id IN (SELECT id FROM yc))
SELECT extract(YEAR from p.date) - ? AS index, SUM(p.value/total.total) AS ratio
FROM payment p, total
WHERE p.financial_commitment_id IN (SELECT id FROM yc) GROUP BY index ORDER BY index`
