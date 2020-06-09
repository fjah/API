package http

import (
	"context"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"slicerapi/internal/config"
	"slicerapi/internal/db"
	"time"
)

func handleInviteJoin(c *gin.Context) {
	userID := jwt.ExtractClaims(c)["id"].(string)
	chID := c.Param("id")

	var channel db.Channel

	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	if err := db.Mongo.Database(config.C.MongoDB.Name).Collection("users").FindOne(
		ctx,
		bson.M{
			"_id": chID,
		},
	).Decode(&channel); err != nil {
		if err == mongo.ErrNoDocuments {
			chk(http.StatusUnauthorized, err, c)
			return
		}
		chk(http.StatusInternalServerError, err, c)
		return
	}

	stat := http.StatusOK

	_, ok := channel.Pending[userID]
	if ok {
		delete(channel.Pending, userID)

		channel.Users[userID] = true
		ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
		if _, err := db.Mongo.Database(config.C.MongoDB.Name).Collection("channels").UpdateOne(
			ctx,
			bson.M{
				"_id": channel.ID,
			},
			bson.D{{
				"$set",
				bson.D{{
					"pending",
					channel.Pending,
				}, {
					"users",
					channel.Users,
				}},
			}},
		); err != nil {
			chk(500, err, c)
			return
		}

		c.JSON(stat, statusMessage{
			Code:    stat,
			Message: "Invite accepted.",
		})
		return
	}

	stat = http.StatusForbidden
	c.JSON(stat, statusMessage{
		Code:    stat,
		Message: "User isn't in the pending list.",
	})
}
