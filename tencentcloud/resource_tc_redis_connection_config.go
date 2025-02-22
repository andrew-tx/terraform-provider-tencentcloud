/*
Provides a resource to create a redis connection_config

Example Usage

Modify the maximum number of connections and maximum network throughput of an instance

```hcl
data "tencentcloud_redis_zone_config" "zone" {
  type_id = 7
}

resource "tencentcloud_vpc" "vpc" {
  cidr_block = "10.0.0.0/16"
  name       = "tf_redis_vpc"
}

resource "tencentcloud_subnet" "subnet" {
  vpc_id            = tencentcloud_vpc.vpc.id
  availability_zone = data.tencentcloud_redis_zone_config.zone.list[0].zone
  name              = "tf_redis_subnet"
  cidr_block        = "10.0.1.0/24"
}

resource "tencentcloud_redis_instance" "foo" {
  availability_zone  = data.tencentcloud_redis_zone_config.zone.list[0].zone
  type_id            = data.tencentcloud_redis_zone_config.zone.list[0].type_id
  password           = "test12345789"
  mem_size           = 8192
  redis_shard_num    = data.tencentcloud_redis_zone_config.zone.list[0].redis_shard_nums[0]
  redis_replicas_num = data.tencentcloud_redis_zone_config.zone.list[0].redis_replicas_nums[0]
  name               = "terrform_test"
  port               = 6379
  vpc_id             = tencentcloud_vpc.vpc.id
  subnet_id          = tencentcloud_subnet.subnet.id
}

resource "tencentcloud_redis_connection_config" "connection_config" {
   instance_id = "crs-fhm9fnv1"
   client_limit = "20000"
   add_bandwidth = "30"
}

```

Import

Redis connectionConfig can be imported, e.g.

```
$ terraform import tencentcloud_redis_connection_config.connection_config instance_id
```
*/
package tencentcloud

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkErrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	redis "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/redis/v20180412"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
)

func resourceTencentCloudRedisConnectionConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudRedisConnectionConfigCreate,
		Read:   resourceTencentCloudRedisConnectionConfigRead,
		Update: resourceTencentCloudRedisConnectionConfigUpdate,
		Delete: resourceTencentCloudRedisConnectionConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "The ID of instance.",
			},

			"client_limit": {
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "The total number of connections per shard.If read-only replicas are not enabled, the lower limit is 10,000 and the upper limit is 40,000.When you enable read-only replicas, the minimum limit is 10,000 and the upper limit is 10,000 * (the number of read replicas +3).",
			},

			"total_bandwidth": {
				Computed:    true,
				Type:        schema.TypeInt,
				Description: "Total bandwidth of the instance = additional bandwidth * number of shards + standard bandwidth * number of shards * (number of primary nodes + number of read-only replica nodes), the number of shards of the standard architecture = 1, in Mb/s.",
			},

			"base_bandwidth": {
				Computed:    true,
				Type:        schema.TypeInt,
				Description: "standard bandwidth. Refers to the bandwidth allocated by the system to each node when an instance is purchased.",
			},

			"add_bandwidth": {
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeInt,
				Description: "Refers to the additional bandwidth of the instance. When the standard bandwidth does not meet the demand, the user can increase the bandwidth by himself. When the read-only copy is enabled, the total bandwidth of the instance = additional bandwidth * number of fragments + standard bandwidth * number of fragments * Max ([number of read-only replicas, 1] ), the number of shards in the standard architecture = 1, and when read-only replicas are not enabled, the total bandwidth of the instance = additional bandwidth * number of shards + standard bandwidth * number of shards, and the number of shards in the standard architecture = 1.",
			},

			"min_add_bandwidth": {
				Computed:    true,
				Type:        schema.TypeInt,
				Description: "Additional bandwidth sets the lower limit.",
			},

			"max_add_bandwidth": {
				Computed:    true,
				Type:        schema.TypeInt,
				Description: "Additional bandwidth is capped.",
			},
		},
	}
}

func resourceTencentCloudRedisConnectionConfigCreate(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_redis_connection_config.create")()
	defer inconsistentCheck(d, meta)()

	var (
		instanceId string
	)
	if v, ok := d.GetOk("instance_id"); ok {
		instanceId = v.(string)
	}

	d.SetId(instanceId)

	return resourceTencentCloudRedisConnectionConfigUpdate(d, meta)
}

func resourceTencentCloudRedisConnectionConfigRead(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_redis_connection_config.read")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)

	ctx := context.WithValue(context.TODO(), logIdKey, logId)

	service := RedisService{client: meta.(*TencentCloudClient).apiV3Conn}

	instanceId := d.Id()

	connectionConfig, err := service.DescribeRedisInstanceById(ctx, instanceId)
	if err != nil {
		return err
	}

	if connectionConfig == nil {
		d.SetId("")
		log.Printf("[WARN]%s resource `RedisConnectionConfig` [%s] not found, please check if it has been deleted.\n", logId, d.Id())
		return nil
	}

	if connectionConfig.InstanceId != nil {
		_ = d.Set("instance_id", connectionConfig.InstanceId)
	}

	if connectionConfig.ClientLimit != nil {
		_ = d.Set("client_limit", connectionConfig.ClientLimit)
	}

	if connectionConfig.NetLimit != nil && connectionConfig.RedisShardNum != nil {
		netLimt := *connectionConfig.NetLimit
		shardNum := *connectionConfig.RedisShardNum
		_ = d.Set("total_bandwidth", netLimt*shardNum*8)
	}

	bandwidthRange, err := service.DescribeBandwidthRangeById(ctx, instanceId)
	if err != nil {
		return err
	}

	if connectionConfig == nil {
		log.Printf("[WARN]%s resource `DescribeBandwidthRangeById` [%s] not found, please check if it has been deleted.\n", logId, d.Id())
		return nil
	}

	if bandwidthRange.BaseBandwidth != nil {
		_ = d.Set("base_bandwidth", bandwidthRange.BaseBandwidth)
	}
	if bandwidthRange.AddBandwidth != nil {
		_ = d.Set("add_bandwidth", bandwidthRange.AddBandwidth)
	}
	if bandwidthRange.MinAddBandwidth != nil {
		_ = d.Set("min_add_bandwidth", bandwidthRange.MinAddBandwidth)
	}
	if bandwidthRange.MaxAddBandwidth != nil {
		_ = d.Set("max_add_bandwidth", bandwidthRange.MaxAddBandwidth)
	}

	return nil
}

func resourceTencentCloudRedisConnectionConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_redis_connection_config.update")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), logIdKey, logId)

	request := redis.NewModifyConnectionConfigRequest()
	response := redis.NewModifyConnectionConfigResponse()

	instanceId := d.Id()
	request.InstanceId = &instanceId

	if v, ok := d.GetOkExists("client_limit"); ok {
		request.ClientLimit = helper.IntInt64(v.(int))
	}

	if v, ok := d.GetOkExists("add_bandwidth"); ok {
		request.Bandwidth = helper.IntInt64(v.(int))
	}

	err := resource.Retry(writeRetryTimeout, func() *resource.RetryError {
		result, e := meta.(*TencentCloudClient).apiV3Conn.UseRedisClient().ModifyConnectionConfig(request)
		if e != nil {
			if ee, ok := e.(*sdkErrors.TencentCloudSDKError); ok {
				if ee.Code == "FailedOperation.SystemError" {
					return resource.NonRetryableError(e)
				}
			}
			return retryError(e)
		} else {
			log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n", logId, request.GetAction(), request.ToJsonString(), result.ToJsonString())
		}
		response = result
		return nil
	})
	if err != nil {
		log.Printf("[CRITAL]%s update redis param failed, reason:%+v", logId, err)
		return err
	}

	service := RedisService{client: meta.(*TencentCloudClient).apiV3Conn}

	taskId := *response.Response.TaskId
	err = resource.Retry(6*readRetryTimeout, func() *resource.RetryError {
		ok, err := service.DescribeTaskInfo(ctx, instanceId, taskId)
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
			return resource.RetryableError(fmt.Errorf("change account is processing"))
		}
	})

	if err != nil {
		log.Printf("[CRITAL]%s redis change connection fail, reason:%s\n", logId, err.Error())
		return err
	}

	return resourceTencentCloudRedisConnectionConfigRead(d, meta)
}

func resourceTencentCloudRedisConnectionConfigDelete(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.tencentcloud_redis_connection_config.delete")()
	defer inconsistentCheck(d, meta)()

	return nil
}
