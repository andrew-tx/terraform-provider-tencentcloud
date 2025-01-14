package tencentcloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTencentCloudNeedFixDcInternetAddressResource_basic(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDcInternetAddress,
				Check:  resource.ComposeTestCheckFunc(resource.TestCheckResourceAttrSet("tencentcloud_dc_internet_address.internet_address", "id")),
			},
			{
				ResourceName:      "tencentcloud_dc_internet_address.internet_address",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const testAccDcInternetAddress = `

resource "tencentcloud_dc_internet_address" "internet_address" {
  mask_len = 30
  addr_type = 2
  addr_proto = 0
}

`
