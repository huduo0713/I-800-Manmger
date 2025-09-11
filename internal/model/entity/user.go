// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// User is the golang structure for table user.
type User struct {
	Id        int         `json:"id"        orm:"id"         description:""` //
	Name      string      `json:"name"      orm:"name"       description:""` //
	Status    int         `json:"status"    orm:"status"     description:""` //
	Age       int         `json:"age"       orm:"age"        description:""` //
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:""` //
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:""` //
}
