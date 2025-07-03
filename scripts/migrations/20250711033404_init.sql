-- Modify "ident_users" table
ALTER TABLE "ident_users" DROP CONSTRAINT "ident_users_core_enums_status", ADD CONSTRAINT "ident_users_core_enums_user_status" FOREIGN KEY ("status_id") REFERENCES "core_enums" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
