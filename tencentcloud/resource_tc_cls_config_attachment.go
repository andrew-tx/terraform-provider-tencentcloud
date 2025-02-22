/*
Provides a resource to create a cls config attachment

Example Usage

```hcl
resource "tencentcloud_cls_config_attachment" "attach" {
  config_id = tencentcloud_cls_config.config.id
  group_id = "27752a9b-9918-440a-8ee7-9c84a14a47ed"
}

Import

cls config_attachment can be imported using the id, e.g.

```
terraform import tencentcloud_cls_config_attachment.attach config_id#group_id
```

```
*/
package tencentcloud

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cls "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cls/v20201016"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
)

func resourceTencentCloudClsConfigAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudClsConfigAttachmentCreate,
		Read:   resourceTencentCloudClsConfigAttachmentRead,
		Delete: resourceTencentCloudClsConfigAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"config_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Collection configuration id.",
			},
			"group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Machine group id.",
			},
		},
	}
}

func resourceTencentCloudClsConfigAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_cls_config_attachment.create")()

	logId := getLogId(contextNil)

	var (
		request  = cls.NewApplyConfigToMachineGroupRequest()
		configId string
		groupId  string
	)

	if v, ok := d.GetOk("config_id"); ok {
		configId = v.(string)
		request.ConfigId = helper.String(configId)

	}
	if v, ok := d.GetOk("group_id"); ok {
		groupId = v.(string)
		request.GroupId = helper.String(groupId)
	}

	err := resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		result, e := meta.(*TencentCloudClient).apiV3Conn.UseClsClient().ApplyConfigToMachineGroup(request)
		if e != nil {
			return retryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n",
				logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
		}
		return nil
	})

	if err != nil {
		return err
	}

	instanceId := helper.IdFormat(configId, groupId)
	d.SetId(instanceId)

	return nil
}

func resourceTencentCloudClsConfigAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_cls_config_attachment.read")()

	idSplit := strings.Split(d.Id(), FILED_SP)
	if len(idSplit) != 2 {
		return fmt.Errorf("id is broken,%s", d.Id())
	}
	configId := idSplit[0]
	groupId := idSplit[1]

	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), logIdKey, logId)
	service := ClsService{client: meta.(*TencentCloudClient).apiV3Conn}

	machineGroup, err := service.DescribeClsMachineGroupByConfigId(ctx, configId, groupId)
	if err != nil {
		return err
	}

	if machineGroup == nil {
		d.SetId("")
		log.Printf("[WARN]%s resource `ClsConfigAttachment` [%s] not found, please check if it has been deleted.\n", logId, d.Id())
		return nil
	}

	_ = d.Set("config_id", configId)
	_ = d.Set("group_id", machineGroup.GroupId)

	return nil
}

func resourceTencentCloudClsConfigAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_cls_config_attachment.delete")()

	logId := getLogId(contextNil)
	request := cls.NewDeleteConfigFromMachineGroupRequest()

	idSplit := strings.Split(d.Id(), FILED_SP)
	if len(idSplit) != 2 {
		return fmt.Errorf("id is broken,%s", d.Id())
	}
	configId := idSplit[0]
	groupId := idSplit[1]
	request.GroupId = helper.String(groupId)
	request.ConfigId = helper.String(configId)

	err := resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		result, e := meta.(*TencentCloudClient).apiV3Conn.UseClsClient().DeleteConfigFromMachineGroup(request)
		if e != nil {
			return retryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n",
				logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
