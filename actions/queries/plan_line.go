package queries

// GetPlanLineAndFinancialDatas fetch all informations about a plan line or the plan lines of a plan including previsions and beneficiary ratios.
// The query is a schema containing parts that must be replaced by calculated requests.
const GetPlanLineAndFinancialDatas = `SELECT p.id, p.name, p.descript, p.value, p.total_value, 
        CAST(fc.value AS bigint) AS commitment, CAST(pr.value AS bigint) AS programmings, :prevQueryPart :beneficiariesQueryPart
      FROM plan_line p
      :beneficiariesCrossQuery
      LEFT JOIN (SELECT f.plan_line_id, SUM(f.value) AS value FROM financial_commitment f
                  WHERE f.plan_line_id NOTNULL AND EXTRACT(year FROM f.date) < :actualYear
                  GROUP BY 1) fc
        ON fc.plan_line_id = p.id
      LEFT JOIN (SELECT op.plan_line_id, SUM(p.value) AS value FROM physical_op op, programmings p 
                  WHERE p.physical_op_id = op.id AND p.year = :actualYear GROUP BY 1) pr 
        ON pr.plan_line_id = p.id
      LEFT JOIN (SELECT * FROM 
        crosstab ('SELECT p.plan_line_id, c.year, SUM(c.value) FROM physical_op p, prev_commitment c 
                    WHERE p.id = c.physical_op_id AND p.plan_line_id NOTNULL GROUP BY 1,2 ORDER BY 1,2',
                  'SELECT m FROM generate_series( :firstYear, :lastYear) AS m')
          AS (plan_line_id INTEGER, :convertQueryPart)) prev 
      ON prev.plan_line_id = p.id
			WHERE :whereQueryPart ORDER BY 1`
