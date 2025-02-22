/*
Use this data source to query detailed information of vpc account_attributes

Example Usage

```hcl
data "tencentcloud_vpc_account_attributes" "account_attributes" {}
```
*/
package tencentcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
)

func dataSourceTencentCloudVpcAccountAttributes() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTencentCloudVpcAccountAttributesRead,
		Schema: map[string]*schema.Schema{
			"account_attribute_set": {
				Computed:    true,
				Type:        schema.TypeList,
				Description: "User account attribute object.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Attribute name.",
						},
						"attribute_values": {
							Type: schema.TypeSet,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Computed:    true,
							Description: "Attribute values.",
						},
					},
				},
			},

			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
		},
	}
}

func dataSourceTencentCloudVpcAccountAttributesRead(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("data_source.tencentcloud_vpc_account_attributes.read")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)

	ctx := context.WithValue(context.TODO(), logIdKey, logId)

	service := VpcService{client: meta.(*TencentCloudClient).apiV3Conn}

	var accountAttributeSet []*vpc.AccountAttribute

	err := resource.Retry(readRetryTimeout, func() *resource.RetryError {
		result, e := service.DescribeVpcAccountAttributes(ctx)
		if e != nil {
			return retryError(e)
		}
		accountAttributeSet = result
		return nil
	})
	if err != nil {
		return err
	}

	ids := make([]string, 0, len(accountAttributeSet))
	tmpList := make([]map[string]interface{}, 0, len(accountAttributeSet))

	if accountAttributeSet != nil {
		for _, accountAttribute := range accountAttributeSet {
			accountAttributeMap := map[string]interface{}{}

			if accountAttribute.AttributeName != nil {
				accountAttributeMap["attribute_name"] = accountAttribute.AttributeName
			}

			if accountAttribute.AttributeValues != nil {
				accountAttributeMap["attribute_values"] = accountAttribute.AttributeValues
			}

			ids = append(ids, *accountAttribute.AttributeName)
			tmpList = append(tmpList, accountAttributeMap)
		}

		_ = d.Set("account_attribute_set", tmpList)
	}

	d.SetId(helper.DataResourceIdsHash(ids))
	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if e := writeToFile(output.(string), tmpList); e != nil {
			return e
		}
	}
	return nil
}
