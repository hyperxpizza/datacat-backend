package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/bootkemp-dev/datacat-backend/database"
	"github.com/bootkemp-dev/datacat-backend/models"
	"github.com/bootkemp-dev/datacat-backend/utils"

	"github.com/gin-gonic/gin"
)

var jobPool models.Pool

func init() {
	jobPool = models.NewPool()
}

func AddJob(c *gin.Context) {
	var request models.NewJobRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if err := utils.ValidateNewJob(request.JobName, request.JobURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	id, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "id not set in context",
		})
		return
	}

	jobID, err := database.InsertNewJob(request.JobName, request.JobURL, request.Frequency, id.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	j := models.NewJob(jobID, id.(int), request.JobName, request.JobURL, request.Frequency)
	jobPool.AddJob(j)
	j.Run()
	c.JSON(http.StatusOK, gin.H{
		"id":   jobID,
		"name": request.JobName,
		"url":  request.JobURL,
	})
	return
}

func GetJobstatus(c *gin.Context) {
	id := c.Param("id")
	jobID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	userID, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "id not set in context",
		})
		return
	}

	job, err := jobPool.GetJob(jobID, userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"status":  job.GetStatus(),
	})

	return
}

func GetJobs(c *gin.Context) {

	userID, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "id not set in context",
		})
		return
	}

	jobIDString := c.Query("id")
	if jobIDString != "" {
		jobID, err := strconv.Atoi(jobIDString)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		job, err := database.GetJobByID(jobID, userID.(int))
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"message": "Job nor found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"job":     &job,
		})
		return
	} else {
		jobs, err := database.GetAllJobsByUserID(userID.(int))
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"message": "Job nor found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"job":     &jobs,
		})
		return
	}

}

func PauseJob(c *gin.Context) {
	id := c.Param("id")
	jobID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	userID, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "id not set in context",
		})
		return
	}

	job, err := jobPool.GetJob(jobID, userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	go job.Stop()
	err = database.UpdateJobActive(false, jobID, job.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.Status(http.StatusOK)
}

func DeleteJob(c *gin.Context) {
	id := c.Param("id")
	jobID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	userID, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "id not set in context",
		})
		return
	}

	//delete job from the pool
	err = jobPool.RemoveJob(jobID, userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// delete job from the database
	err = database.DeleteJob(jobID, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.Status(http.StatusOK)
	return
}

func RestartJob(c *gin.Context) {
	id := c.Param("id")
	jobID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	userID, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "id not set in context",
		})
		return
	}

	//get job
	job, err := jobPool.GetJob(jobID, userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if job.Active == true {
		job.Stop()
		job.Run()
	} else {
		job.Run()
	}
}

func GetJobActive(c *gin.Context) {
	id := c.Param("id")
	jobID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	userID, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "id not set in context",
		})
		return
	}

	//get job
	job, err := jobPool.GetJob(jobID, userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"active":  job.Active,
	})
}

func JobInfoWebsocket(c *gin.Context) {
	handler := websocket.Handler
}
