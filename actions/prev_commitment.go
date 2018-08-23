package actions

import (
	"net/http"

	"github.com/Iledant/iris_propera/models"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

// prevCommitmentReq is used to decode one line of previsional commitment batch
type prevCommitmentReq struct {
	Number     string             `json:"number"`
	Year       int64              `json:"year"`
	Value      int64              `json:"value"`
	TotalValue models.NullInt64   `json:"total_value"`
	StateRatio models.NullFloat64 `json:"state_ratio"`
}

// batchPrevCommitment is used to decode batch payload
type batchPrevCommitment struct {
	Pcs []prevCommitmentReq `json:"PrevCommitment"`
}

// BatchPrevCommitments handles the post request to upload an array of previsional commitments
func BatchPrevCommitments(ctx iris.Context) {
	req := batchPrevCommitment{}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch prévision d'engagements : erreur décodage " + err.Error()})
		return
	}

	db := ctx.Values().Get("db").(*gorm.DB)
	tx := db.Begin()
	if err := tx.Exec("DROP TABLE IF EXISTS temp_prev_commitment").Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch prévision d'engagements, suppression table temporaire : " + err.Error()})
		tx.Rollback()
		return
	}
	if err := tx.Exec(`CREATE TABLE temp_prev_commitment (number varchar(10), year integer, value bigint,
	 total_value bigint, state_ratio double precision)`).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch prévision d'engagements, création table temporaire : " + err.Error()})
		tx.Rollback()
		return
	}

	for _, pc := range req.Pcs {
		if err := tx.Exec(`INSERT INTO temp_prev_commitment (number,year,value,total_value,state_ratio) VALUES(?,?,?,?,?)`,
			pc.Number, pc.Year, pc.Value, pc.TotalValue, pc.StateRatio).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(jsonError{"Batch prévision d'engagements, insertion table temporaire : " + err.Error()})
			tx.Rollback()
			return
		}
	}
	if err := tx.Exec(`UPDATE prev_commitment SET value = t.value, total_value = t.total_value, 
	state_ratio = t.state_ratio FROM temp_prev_commitment t, physical_op op
	WHERE t.number=op.number AND prev_commitment.physical_op_id = op.id AND t.year = prev_commitment.year`).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch prévision d'engagements, update : " + err.Error()})
		tx.Rollback()
		return
	}
	if err := tx.Exec(`INSERT INTO prev_commitment (physical_op_id, year, value, descript, total_value, state_ratio)
	SELECT op.id, t.year, t.value, NULL, t.total_value, t.state_ratio FROM physical_op op, temp_prev_commitment t
	WHERE op.number = t.number AND ((op.id, t.year) NOT IN (SELECT physical_op_id, year FROM prev_commitment))`).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch prévision d'engagements, insert : " + err.Error()})
		tx.Rollback()
		return
	}
	if err := tx.Exec("DROP TABLE IF EXISTS temp_prev_commitment").Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Batch prévision d'engagements, suppression table temporaire : " + err.Error()})
		tx.Rollback()
		return
	}
	tx.Commit()
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Batch prévision d'engagement importé"})
}
