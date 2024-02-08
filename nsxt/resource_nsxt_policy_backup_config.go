/* Copyright © 2023 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"log"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/cluster/backups"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

const defaultBackupSite = "localhost"

var schemeNameValues = []string{
	model.FileTransferAuthenticationScheme_SCHEME_NAME_PASSWORD,
	model.FileTransferAuthenticationScheme_SCHEME_NAME_KEY,
}

func resourceNsxtPolicyBackupConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyBackupConfigCreate,
		Read:   resourceNsxtPolicyBackupConfigRead,
		Update: resourceNsxtPolicyBackupConfigUpdate,
		Delete: resourceNsxtPolicyBackupConfigDelete,

		Schema: map[string]*schema.Schema{
			"after_inventory_update_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "A number of seconds after a last backup, that needs to pass, before a topology change will trigger a generation of a new cluster/node backups",
				ValidateFunc: validation.IntBetween(300, 86400),
			},
			"backup_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "true if automated backup is enabled",
			},
			"interval_backup_schedule": {
				ExactlyOneOf: []string{"weekly_backup_schedule"},
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				Description:  "Schedule to specify the interval time at which automated backups need to be taken",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"seconds_between_backups": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     3600,
							Description: "Time interval in seconds between two consecutive automated backups",
						},
					},
				},
			},
			"weekly_backup_schedule": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Schedule to specify day of the week and time to take automated backup",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"days_of_week": {
							Type:        schema.TypeList,
							Required:    true,
							Description: "Days of week when backup is taken. 0 - Sunday, 1 - Monday, 2 - Tuesday, 3 - Wednesday ...",
							MinItems:    1,
							MaxItems:    7,
							Elem: &schema.Schema{
								Type:         schema.TypeInt,
								ValidateFunc: validation.IntBetween(0, 6),
							},
						},
						"hour_of_day": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Time of day when backup is taken",
							ValidateFunc: validation.IntBetween(0, 23),
						},
						"minute_of_day": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Time of day when backup is taken",
							ValidateFunc: validation.IntBetween(0, 59),
						},
					},
				},
			},
			"inventory_summary_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "The minimum number of seconds between each upload of the inventory summary to backup server",
				ValidateFunc: validation.IntBetween(30, 3600),
				Default:      240,
			},
			"passphrase": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Passphrase used to encrypt backup files",
			},
			"remote_file_server": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "The server to which backups will be sent",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"directory_path": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Remote server directory to copy bundle files to",
							ValidateFunc: validation.StringMatch(regexp.MustCompile("^\\/[\\w\\-.\\+~\\/]+$"),
								"Directory path doesn't match required pattern ^\\/[\\w\\-.\\+~\\/]+$"),
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Server port",
							Default:      22,
							ValidateFunc: validateSinglePort(),
						},
						"protocol": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Required:    true,
							Description: "Protocol to use to copy file",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authentication_scheme": {
										Type:        schema.TypeList,
										MaxItems:    1,
										Required:    true,
										Description: "Scheme to authenticate if required",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"identity_file": {
													Type:        schema.TypeString,
													Optional:    true,
													Sensitive:   true,
													Description: "SSH private key data",
												},
												"password": {
													Type:        schema.TypeString,
													Optional:    true,
													Sensitive:   true,
													Description: "Password to authenticate with",
												},
												"scheme_name": {
													Type:         schema.TypeString,
													Required:     true,
													Description:  "Authentication scheme name",
													ValidateFunc: validation.StringInSlice(schemeNameValues, false),
												},
												"username": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "User name to authenticate with",
												},
											},
										},
									},
									"protocol_name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Protocol name",
										Default:     "sftp",
									},
									"ssh_fingerprint": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "SSH fingerprint of server",
										ValidateFunc: validation.StringMatch(regexp.MustCompile("^SHA256:.*$"),
											"SSH fingerprint doesn't match pattern ^SHA256:.*$"),
									},
								},
							},
						},
						"server": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Remote server hostname or IP address",
						},
					},
				},
			},
		},
	}
}

func resourceNsxtPolicyBackupConfigCreate(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicyBackupConfigUpdate(d, m)
}

func getRemoteFileServerFromSchema(rfsList interface{}) *model.RemoteFileServer {
	for _, r := range rfsList.([]interface{}) {
		rfs := r.(map[string]interface{})
		directoryPath := rfs["directory_path"].(string)
		port := int64(rfs["port"].(int))
		server := rfs["server"].(string)
		var protocol *model.FileTransferProtocol
		for _, p := range rfs["protocol"].([]interface{}) {
			proto := p.(map[string]interface{})
			var authScheme *model.FileTransferAuthenticationScheme
			for _, as := range proto["authentication_scheme"].([]interface{}) {
				aScheme := as.(map[string]interface{})
				identityFile := aScheme["identity_file"].(string)
				password := aScheme["password"].(string)
				schemeName := aScheme["scheme_name"].(string)
				username := aScheme["username"].(string)
				authScheme = &model.FileTransferAuthenticationScheme{
					IdentityFile: &identityFile,
					Password:     &password,
					SchemeName:   &schemeName,
					Username:     &username,
				}
			}
			protocolName := proto["protocol_name"].(string)
			sshFingerprint := proto["ssh_fingerprint"].(string)
			protocol = &model.FileTransferProtocol{
				AuthenticationScheme: authScheme,
				ProtocolName:         &protocolName,
				SshFingerprint:       &sshFingerprint,
			}
		}

		return &model.RemoteFileServer{
			DirectoryPath: &directoryPath,
			Port:          &port,
			Protocol:      protocol,
			Server:        &server,
		}
	}
	return nil
}

func getBackupScheduleFromSchema(d *schema.ResourceData) (*data.StructValue, error) {
	converter := bindings.NewTypeConverter()
	var dataValue data.DataValue
	var errs []error

	for _, scheduleList := range []string{"interval_backup_schedule", "weekly_backup_schedule"} {
		for _, s := range d.Get(scheduleList).([]interface{}) {
			schedule := s.(map[string]interface{})
			switch scheduleList {
			case "interval_backup_schedule":
				secondsBetweenBackups := int64(schedule["seconds_between_backups"].(int))
				obj := model.IntervalBackupSchedule{
					SecondsBetweenBackups: &secondsBetweenBackups,
					ResourceType:          model.IntervalBackupSchedule__TYPE_IDENTIFIER,
				}
				dataValue, errs = converter.ConvertToVapi(obj, model.IntervalBackupScheduleBindingType())

			case "weekly_backup_schedule":
				daysOfWeek := intList2int64List(schedule["days_of_week"].([]interface{}))
				hourOfDay := int64(schedule["hour_of_day"].(int))
				minuteOfDay := int64(schedule["minute_of_day"].(int))
				obj := model.WeeklyBackupSchedule{
					DaysOfWeek:   daysOfWeek,
					HourOfDay:    &hourOfDay,
					MinuteOfDay:  &minuteOfDay,
					ResourceType: model.WeeklyBackupSchedule__TYPE_IDENTIFIER,
				}
				dataValue, errs = converter.ConvertToVapi(obj, model.WeeklyBackupScheduleBindingType())
			}
		}
	}
	if errs != nil {
		log.Printf("Failed to convert schedule object, errors are %v", errs)
		return nil, errs[0]
	} else if dataValue != nil {
		return dataValue.(*data.StructValue), nil
	}
	return nil, nil
}

func resourceNsxtPolicyBackupConfigRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := backups.NewConfigClient(connector)
	obj, err := client.Get()
	if err != nil {
		return handleReadError(d, "PolicyBackupConfig", "", err)
	}
	if obj.AfterInventoryUpdateInterval != nil {
		d.Set("after_inventory_update_interval", obj.AfterInventoryUpdateInterval)
	}
	d.Set("backup_enabled", obj.BackupEnabled)
	err = setBackupScheduleInSchema(d, obj.BackupSchedule)
	if err != nil {
		return handleReadError(d, "PolicyBackupConfig", "", err)
	}
	d.Set("inventory_summary_interval", obj.InventorySummaryInterval)
	d.Set("passphrase", obj.Passphrase)
	var rfs map[string]interface{}
	if obj.RemoteFileServer != nil {
		rfs = getElemOrEmptyMapFromSchema(d, "remote_file_server")
		rfs["directory_path"] = obj.RemoteFileServer.DirectoryPath
		rfs["port"] = obj.RemoteFileServer.Port
		var protocol map[string]interface{}
		if obj.RemoteFileServer.Protocol != nil {
			protocol = getElemOrEmptyMapFromMap(rfs, "protocol")
			var authScheme map[string]interface{}
			if obj.RemoteFileServer.Protocol.AuthenticationScheme != nil {
				authScheme = getElemOrEmptyMapFromMap(protocol, "authentication_scheme")
				authScheme["identity_file"] = obj.RemoteFileServer.Protocol.AuthenticationScheme.IdentityFile
				if obj.RemoteFileServer.Protocol.AuthenticationScheme.Password != nil {
					authScheme["password"] = obj.RemoteFileServer.Protocol.AuthenticationScheme.Password
				}
				authScheme["scheme_name"] = obj.RemoteFileServer.Protocol.AuthenticationScheme.SchemeName
				authScheme["username"] = obj.RemoteFileServer.Protocol.AuthenticationScheme.Username
			}
			protocol["authentication_scheme"] = []interface{}{authScheme}
			protocol["protocol_name"] = obj.RemoteFileServer.Protocol.ProtocolName
			protocol["ssh_fingerprint"] = obj.RemoteFileServer.Protocol.SshFingerprint
		}
		rfs["protocol"] = []interface{}{protocol}
		rfs["server"] = obj.RemoteFileServer.Server
	}
	d.Set("remote_file_server", []interface{}{rfs})
	return nil
}

func setBackupScheduleInSchema(d *schema.ResourceData, schedule *data.StructValue) error {
	converter := bindings.NewTypeConverter()
	base, errs := converter.ConvertToGolang(schedule, model.BackupScheduleBindingType())
	if errs != nil {
		return errs[0]
	}
	scheduleType := base.(model.BackupSchedule).ResourceType

	switch scheduleType {
	case model.IntervalBackupSchedule__TYPE_IDENTIFIER:
		schemaSched := make(map[string]interface{})
		s, errs := converter.ConvertToGolang(schedule, model.IntervalBackupScheduleBindingType())
		if errs != nil {
			return errs[0]
		}
		sched := s.(model.IntervalBackupSchedule)
		secondsBetweenBackups := sched.SecondsBetweenBackups
		schemaSched["seconds_between_backups"] = secondsBetweenBackups
		d.Set("interval_backup_schedule", []interface{}{schemaSched})

	case model.WeeklyBackupSchedule__TYPE_IDENTIFIER:
		schemaSched := make(map[string]interface{})
		s, errs := converter.ConvertToGolang(schedule, model.WeeklyBackupScheduleBindingType())
		if errs != nil {
			return errs[0]
		}
		sched := s.(model.WeeklyBackupSchedule)
		schemaSched["days_of_week"] = sched.DaysOfWeek
		schemaSched["hour_of_day"] = sched.HourOfDay
		schemaSched["minute_of_day"] = sched.MinuteOfDay
		d.Set("weekly_backup_schedule", []interface{}{schemaSched})

	default:
		return fmt.Errorf("unknown schedule type %s", scheduleType)
	}
	return nil
}

func resourceNsxtPolicyBackupConfigUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := backups.NewConfigClient(connector)

	siteID := defaultBackupSite
	frameType := backups.Config_UPDATE_FRAME_TYPE_LOCAL_LOCAL_MANAGER
	if isPolicyGlobalManager(m) {
		frameType = backups.Config_UPDATE_FRAME_TYPE_GLOBAL_MANAGER
	}
	afterInventoryUpdateInterval := int64(d.Get("after_inventory_update_interval").(int))
	backupEnabled := d.Get("backup_enabled").(bool)
	backupSchedule, err := getBackupScheduleFromSchema(d)
	if err != nil {
		return handleUpdateError("PolicyBackupConfig", "", err)
	}
	inventorySummaryInterval := int64(d.Get("inventory_summary_interval").(int))
	passphrase := d.Get("passphrase").(string)
	remoteFileServer := getRemoteFileServerFromSchema(d.Get("remote_file_server"))

	obj := model.BackupConfiguration{
		BackupEnabled:            &backupEnabled,
		InventorySummaryInterval: &inventorySummaryInterval,
		Passphrase:               &passphrase,
		BackupSchedule:           backupSchedule,
		RemoteFileServer:         remoteFileServer,
	}
	if afterInventoryUpdateInterval != 0 {
		obj.AfterInventoryUpdateInterval = &afterInventoryUpdateInterval
	}
	_, err = client.Update(obj, &frameType, &siteID)
	if err != nil {
		return handleCreateError("PolicyBackupConfig", "", err)
	}

	d.SetId(siteID)
	return resourceNsxtPolicyBackupConfigRead(d, m)
}

func resourceNsxtPolicyBackupConfigDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := backups.NewConfigClient(connector)

	siteID := defaultBackupSite
	frameType := backups.Config_UPDATE_FRAME_TYPE_LOCAL_LOCAL_MANAGER
	if isPolicyGlobalManager(m) {
		frameType = backups.Config_UPDATE_FRAME_TYPE_GLOBAL_MANAGER
	}
	obj := model.BackupConfiguration{}
	_, err := client.Update(obj, &frameType, &siteID)
	if err != nil {
		return handleDeleteError("PolicyBackupConfig", "", err)
	}
	return nil
}
