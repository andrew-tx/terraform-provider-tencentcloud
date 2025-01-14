/*
Provides a resource to create a cynosdb wan

Example Usage

```hcl
resource "tencentcloud_cynosdb_wan" "wan" {
  cluster_id      = "cynosdbmysql-bws8h88b"
  instance_grp_id = "cynosdbmysql-grp-lxav0p9z"
}
```

Import

cynosdb wan can be imported using the id, e.g.

```
terraform import tencentcloud_cynosdb_wan.wan cynosdbmysql-bws8h88b#cynosdbmysql-grp-lxav0p9z
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
	sdkErrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	cynosdb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cynosdb/v20190107"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
)

func resourceTencentCloudCynosdbWan() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudCynosdbWanCreate,
		Read:   resourceTencentCloudCynosdbWanRead,
		Update: resourceTencentCloudCynosdbWanUpdate,
		Delete: resourceTencentCloudCynosdbWanDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "Cluster ID.",
			},

			"instance_grp_id": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "Instance Group ID.",
			},

			"wan_domain": {
				Computed:    true,
				Type:        schema.TypeString,
				Description: "Domain name.",
			},

			"wan_ip": {
				Computed:    true,
				Type:        schema.TypeString,
				Description: "Network ip.",
			},

			"wan_port": {
				Computed:    true,
				Type:        schema.TypeInt,
				Description: "Internet port.",
			},

			"wan_status": {
				Computed:    true,
				Type:        schema.TypeString,
				Description: "Internet status.",
			},
		},
	}
}

func resourceTencentCloudCynosdbWanCreate(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_cynosdb_wan.create")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), logIdKey, logId)

	var (
		request       = cynosdb.NewOpenWanRequest()
		response      = cynosdb.NewOpenWanResponse()
		clusterId     string
		instanceGrpId string
	)

	if v, ok := d.GetOk("cluster_id"); ok {
		clusterId = v.(string)
	}

	if v, ok := d.GetOk("instance_grp_id"); ok {
		instanceGrpId = v.(string)
		request.InstanceGrpId = helper.String(v.(string))
	}

	err := resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		result, e := meta.(*TencentCloudClient).apiV3Conn.UseCynosdbClient().OpenWan(request)
		if e != nil {
			return retryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
		}
		response = result
		return nil
	})
	if err != nil {
		log.Printf("[CRITAL]%s Open cynosdb wan failed, reason:%+v", logId, err)
		return err
	}

	d.SetId(clusterId + FILED_SP + instanceGrpId)

	flowId := *response.Response.FlowId
	service := CynosdbService{client: meta.(*TencentCloudClient).apiV3Conn}
	err = resource.Retry(6*readRetryTimeout, func() *resource.RetryError {
		ok, err := service.DescribeFlow(ctx, flowId)
		if err != nil {
			if _, ok := err.(*sdkErrors.TencentCloudSDKError); !ok {
				return resource.RetryableError(err)
			} else {
				return resource.NonRetryableError(err)
			}
		}
		if ok {
			return nil
		} else {
			return resource.RetryableError(fmt.Errorf("Open cynosdb wan is processing"))
		}
	})

	if err != nil {
		log.Printf("[CRITAL]%s Open cynosdb wan fail, reason:%s\n", logId, err.Error())
		return err
	}

	return resourceTencentCloudCynosdbWanRead(d, meta)
}

func resourceTencentCloudCynosdbWanRead(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_cynosdb_wan.read")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), logIdKey, logId)

	service := CynosdbService{client: meta.(*TencentCloudClient).apiV3Conn}

	idSplit := strings.Split(d.Id(), FILED_SP)
	if len(idSplit) != 2 {
		return fmt.Errorf("id is broken,%s", d.Id())
	}
	clusterId := idSplit[0]
	instanceGrpId := idSplit[1]

	wan, err := service.DescribeClusterInstanceGrps(ctx, clusterId)
	if err != nil {
		return err
	}

	if wan == nil {
		d.SetId("")
		log.Printf("[WARN]%s resource `CynosdbWan` [%s] not found, please check if it has been deleted.\n", logId, d.Id())
		return nil
	}

	_ = d.Set("cluster_id", clusterId)
	_ = d.Set("instance_grp_id", instanceGrpId)

	if wan.Response != nil && wan.Response.InstanceGrpInfoList != nil {
		for _, v := range wan.Response.InstanceGrpInfoList {
			if *v.InstanceGrpId == instanceGrpId {
				if v.WanDomain != nil {
					_ = d.Set("wan_domain", v.WanDomain)
				}
				if v.WanIP != nil {
					_ = d.Set("wan_ip", v.WanIP)
				}
				if v.WanPort != nil {
					_ = d.Set("wan_port", v.WanPort)
				}
				if v.WanStatus != nil {
					_ = d.Set("wan_status", v.WanStatus)
				}
			}
		}
	}

	return nil
}

func resourceTencentCloudCynosdbWanUpdate(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_cynosdb_wan.update")()
	defer inconsistentCheck(d, meta)()

	immutableArgs := []string{"cluster_id", "instance_grp_id"}

	for _, v := range immutableArgs {
		if d.HasChange(v) {
			return fmt.Errorf("argument `%s` cannot be changed", v)
		}
	}
	return resourceTencentCloudCynosdbWanRead(d, meta)
}

func resourceTencentCloudCynosdbWanDelete(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_cynosdb_wan.delete")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), logIdKey, logId)

	service := CynosdbService{client: meta.(*TencentCloudClient).apiV3Conn}
	idSplit := strings.Split(d.Id(), FILED_SP)
	if len(idSplit) != 2 {
		return fmt.Errorf("id is broken,%s", d.Id())
	}
	// clusterId := idSplit[0]
	instanceGrpId := idSplit[1]

	flowId, err := service.DeleteCynosdbWanById(ctx, instanceGrpId)
	if err != nil {
		return err
	}

	err = resource.Retry(6*readRetryTimeout, func() *resource.RetryError {
		ok, err := service.DescribeFlow(ctx, flowId)
		if err != nil {
			if _, ok := err.(*sdkErrors.TencentCloudSDKError); !ok {
				return resource.RetryableError(err)
			} else {
				return resource.NonRetryableError(err)
			}
		}
		if ok {
			return nil
		} else {
			return resource.RetryableError(fmt.Errorf("Close cynosdb wan is processing"))
		}
	})

	if err != nil {
		log.Printf("[CRITAL]%s Close cynosdb wan fail, reason:%s\n", logId, err.Error())
		return err
	}

	return nil
}
