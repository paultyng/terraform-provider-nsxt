// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	nsxModel "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/node"
)

func resourceNsxtUsers() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtNodeUserCreate,
		Read:   resourceNsxtNodeUserRead,
		Update: resourceNsxtNodeUserUpdate,
		Delete: resourceNsxtNodeUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"full_name": {
				Type:        schema.TypeString,
				Description: "Full name for the user",
				Required:    true,
			},
			"last_password_change": {
				Type:        schema.TypeInt,
				Description: "Number of days since password was last changed",
				Computed:    true,
			},
			"password": {
				Type:         schema.TypeString,
				Description:  "Password for the user",
				Sensitive:    true,
				Optional:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"password_change_frequency": {
				Type:         schema.TypeInt,
				Description:  "Number of days password is valid before it must be changed",
				Optional:     true,
				Default:      90,
				ValidateFunc: validation.IntBetween(0, 9999),
			},
			"password_change_warning": {
				Type:         schema.TypeInt,
				Description:  "Number of days before user receives warning message of password expiration",
				Optional:     true,
				Default:      7,
				ValidateFunc: validation.IntBetween(0, 9999),
			},
			"password_reset_required": {
				Type:        schema.TypeBool,
				Description: "Boolean value that states if a password reset is required",
				Computed:    true,
			},
			"active": {
				Type:        schema.TypeBool,
				Description: "Boolean value that states if the user account is activated",
				Optional:    true,
				Default:     true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "User status",
				Computed:    true,
			},
			"user_id": {
				Type:        schema.TypeInt,
				Description: "Numeric id for the user",
				Computed:    true,
			},
			"username": {
				Type:         schema.TypeString,
				Description:  "User login name",
				ValidateFunc: validateUsername(),
				Required:     true,
			},
		},
	}
}

func validateUsername() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		r := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9@-_.\-]*$`)
		if ok := r.MatchString(v); !ok {
			es = append(es, fmt.Errorf("username %s is invalid", v))
			return
		}

		if len(v) < 1 || len(v) > 32 {
			es = append(es, fmt.Errorf("expected length of %s to be in the range (%d - %d), got %d", k, 1, 32, len(v)))
			return
		}
		return
	}
}

func resourceNsxtNodeUserCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := node.NewUsersClient(connector)

	fullName := d.Get("full_name").(string)
	passwordChangeFrequency := int64(d.Get("password_change_frequency").(int))
	passwordChangeWarning := int64(d.Get("password_change_warning").(int))
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	userProp := nsxModel.NodeUserProperties{
		FullName:                &fullName,
		PasswordChangeFrequency: &passwordChangeFrequency,
		PasswordChangeWarning:   &passwordChangeWarning,
		Username:                &username,
	}
	if len(password) > 0 {
		userProp.Password = &password
	}

	user, err := client.Createuser(userProp)
	if err != nil {
		return handleCreateError("User", username, err)
	}
	d.Set("user_id", user.Userid)
	d.SetId(strconv.Itoa(int(*user.Userid)))

	return resourceNsxtNodeUserRead(d, m)
}

func resourceNsxtNodeUserRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := node.NewUsersClient(connector)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining logical object id")
	}
	user, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "User", id, err)
	}

	// Password not return on GET
	d.Set("full_name", user.FullName)
	d.Set("last_password_change", user.LastPasswordChange)
	d.Set("password_change_frequency", user.PasswordChangeFrequency)
	d.Set("password_change_warning", user.PasswordChangeWarning)
	d.Set("password_reset_required", user.PasswordResetRequired)
	d.Set("status", user.Status)
	d.Set("user_id", user.Userid)
	d.Set("username", user.Username)
	if *user.Status == nsxModel.NodeUserProperties_STATUS_NOT_ACTIVATED {
		d.Set("active", false)
	} else {
		d.Set("active", true)
	}
	return nil
}

func resourceNsxtNodeUserUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := node.NewUsersClient(connector)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining logical object id")
	}
	oldPwd, pwd := d.GetChange("password")
	password := pwd.(string)
	oldPassword := oldPwd.(string)

	active := d.Get("active").(bool)
	status := d.Get("status").(string)

	// Handle user status change first
	// Password reset can be achieved by deactivating then re-activate the account
	if status == nsxModel.NodeUserProperties_STATUS_NOT_ACTIVATED && active {
		if len(password) == 0 {
			return fmt.Errorf("must specify password to activate Nsxt Node user %s", id)
		}
		_, err := client.Activate(id, nsxModel.NodeUserPasswordProperty{Password: &password})
		if err != nil {
			return fmt.Errorf("failed to activate Nsxt Node user %s: %s", id, err)
		}
	} else if status != nsxModel.NodeUserProperties_STATUS_NOT_ACTIVATED && !active {
		_, err := client.Deactivate(id)
		if err != nil {
			return fmt.Errorf("failed to deactivate Nsxt Node user %s: %s", id, err)
		}
	}

	// Update other user properties
	fullName := d.Get("full_name").(string)
	passwordChangeFrequency := int64(d.Get("password_change_frequency").(int))
	passwordChangeWarning := int64(d.Get("password_change_warning").(int))
	username := d.Get("username").(string)
	userProp := nsxModel.NodeUserProperties{
		FullName:                &fullName,
		PasswordChangeFrequency: &passwordChangeFrequency,
		PasswordChangeWarning:   &passwordChangeWarning,
		Username:                &username,
	}

	// If password is changed, handle password change.
	if password != oldPassword {
		userProp.Password = &password
		userProp.OldPassword = &oldPassword
	}

	_, err := client.Update(id, userProp)
	if err != nil {
		return handleUpdateError("User", id, err)
	}
	return resourceNsxtNodeUserRead(d, m)
}

func resourceNsxtNodeUserDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := node.NewUsersClient(connector)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining logical object id")
	}

	// Workaround a bug where the delete response body is marked as non-json content type by NSX,
	// triggering SDK to raise it as internal server error.
	// Error Data is not complying with ApiErrorBindingType, making the conversion result empty.
	// A second deletion attempt should be deemed as NotFound, which is then treated as successful.
	var err error
	for i := 0; i < 2; i++ {
		err = client.Delete(id)
		if err == nil {
			return nil
		}
		err = handleDeleteError("User", id, err)
		time.Sleep(1 * time.Second)
	}

	return err
}
