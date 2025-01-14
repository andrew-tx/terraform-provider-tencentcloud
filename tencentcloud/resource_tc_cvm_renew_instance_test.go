package tencentcloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTencentCloudCvmRenewInstanceResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheckCommon(t, ACCOUNT_TYPE_PREPAY) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCvmRenewInstance,
				Check:  resource.ComposeTestCheckFunc(resource.TestCheckResourceAttrSet("tencentcloud_cvm_renew_instance.renew_instance", "id")),
			},
		},
	})
}

const testAccCvmRenewInstance = testAccTencentCloudInstanceBasicToPrepaid + `

resource "tencentcloud_cvm_renew_instance" "renew_instance" {
  instance_id = tencentcloud_instance.foo.id
  instance_charge_prepaid {
	period = 1
  }
}

`
