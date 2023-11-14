package main

import (
	"log"
	"time"

	"github.com/Lukman-e/try-gost/database/connector"
	"github.com/Lukman-e/try-gost/domain/entity"
	"github.com/Lukman-e/try-gost/internal/env"
	"github.com/Lukman-e/try-gost/internal/rbac"
	"gorm.io/gorm"
)

var (
	db     *gorm.DB
	config env.Config
)

func setup() {
	db = connector.LoadDatabase()
	env.ReadConfig("./.env")
	config = env.Configuration()
}

// ⚠️ Do not run this on production.
// ⚠️ Warning: This script will drop all tables in Database.
// This script is designed to perform a complete reset of the database,
// which involves the deletion of all existing tables and recreating them from scratch.
func main() {
	setup()
	log.Println("Start Migration")
	defer log.Println("Finish Migration: success with no error")

	// delete all tables if not in production
	if !config.GetAppInProduction() {
		dropAll()
	}

	// Create a new transaction
	tx := db.Begin()
	if tx.Error != nil {
		log.Panicf("Error starting transaction: %s", tx.Error)
	}

	// Recreate or create new tables within the transaction
	if migrateErr := tx.AutoMigrate(entity.AllTables()...); migrateErr != nil {
		tx.Rollback()
		log.Panicf("Error while migrating DB : %s", migrateErr)
	}

	// Commit the transaction if not in production
	if !config.GetAppInProduction() {
		if commitErr := tx.Commit().Error; commitErr != nil {
			tx.Rollback()
			log.Panicf("Error committing transaction: %s", commitErr)
		}
	}

	// Seed master-RBAC data (roles and permissions)
	if !config.GetAppInProduction() {
		seeding()
	}
}

// dropAll func drops all tables that listed in entity.AllTables()
func dropAll() {
	log.Println("Warning: dropping all tables in 9 seconds (CTRL+C to stop)")
	time.Sleep(10 * time.Second)
	log.Println("Start dropping tables . . .")
	tables := entity.AllTables()
	if deleteErr := db.Migrator().DropTable(tables...); deleteErr != nil {
		log.Panicf("Error while deleting tables DB: %s", deleteErr)
	}
}

// seeding func seed data like role, permission,
// and role_has_permissions tables. You can add
// more seed if you want.
func seeding() {
	// Create a new transaction for seeding
	tx := db.Begin()
	if tx.Error != nil {
		log.Panicf("Error starting transaction for seeding: %s", tx.Error)
	}

	// Seeding permission and role
	for _, data := range rbac.AllRoles() {
		if createErr := tx.Create(&data).Error; createErr != nil {
			tx.Rollback()
			log.Panicf("Error while creating Roles: %s", createErr)
		}
	}
	for _, perm := range rbac.AllPermissions() {
		perm.SetCreateTime()
		perm.ID = 0
		if createErr := tx.Create(&perm).Error; createErr != nil {
			tx.Rollback()
			log.Panicf("Error while creating Permissions: %s", createErr)
		}

		if perm.ID <= 20 {
			if createErr := tx.Create(&entity.RoleHasPermission{
				RoleID:       1, // admin
				PermissionID: perm.ID,
			}).Error; createErr != nil {
				tx.Rollback()
				log.Panicf("Error while creating Roles: %s", createErr)
			}
		}
		if perm.ID > 10 {
			if createErr := tx.Create(&entity.RoleHasPermission{
				RoleID:       2, // user
				PermissionID: perm.ID,
			}).Error; createErr != nil {
				tx.Rollback()
				log.Panicf("Error while creating Roles: %s", createErr)
			}
		}
	}

	// Commit the transaction for seeding
	if commitErr := tx.Commit().Error; commitErr != nil {
		tx.Rollback()
		log.Panicf("Error committing transaction for seeding: %s", commitErr)
	}
}
