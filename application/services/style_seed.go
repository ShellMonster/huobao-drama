package services

type StyleSeed struct {
	Name     string
	ImageURL string
}

func DefaultStyleSeeds() []StyleSeed {
	return []StyleSeed{
		{
			Name:     "日漫异世界风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/ri_man_yi_shi_jie.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768984099%3B1769588899%26q-key-time%3D1768984099%3B1769588899%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3De32d9723d7e412124c1bcfb0690f2cd4516f319c&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "奇幻卡通风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/qi_huan_ka_tong.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768984099%3B1769588899%26q-key-time%3D1768984099%3B1769588899%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D7cf50f12a87fdeedb10cbf5e091cf854ba268864&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "国漫仙侠风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/guo_man_xian_xia.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768984099%3B1769588899%26q-key-time%3D1768984099%3B1769588899%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3Db8a7740cea05375ff8cc62527eaa9afdeed23ecf&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "CG史诗风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/cgwrzs.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D68a9b8771a4215c98cda359e0b2b7882935f665b&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "极细线条韩漫",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/ylqts.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D0789b3738d2d23cc092ebc551859db0e48942f69&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "韩漫古装风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/zqfm.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3Dbcf4269cfd71152ab20e1f55f32f8bd9cb2dcff3&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "韩漫都市风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/hmds.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D40ede0c347c2cd64c2e59d54a7efbe41782f4a62&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "动漫通用风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/dmty.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3Df9abebb9879223311fe67a7e0f1c3ee3a4810b80&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "古风水墨风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/gfsm.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D06c5d76995bb5f9d00b9108a2f4d22b13c158140&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "国风卡通风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/guo_man_ka_tong.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768984099%3B1769588899%26q-key-time%3D1768984099%3B1769588899%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3Dec613015ef9bf6f052fda8473f20e6ac0f8bfabe&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "都市动漫风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/wdeq.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D322ce2e2fae9bc634d85dea0996e9f02232da1f7&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "校园卡通风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/npxy.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3Def2f6451e05b656095811346572f18ec9f2996d1&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "CG都市风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/cgds.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D910919da27bb69aad72e2b70dab5d42cd39b2418&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "美漫风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/mm.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D177d8f4ce94cc5333052022d4345facbb37409ec&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "2d平面风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/2dpm.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D18757cc57fa900dfe60ca5a1c03f55bacea767eb&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "唯美光影二次元",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/wmgyecy.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1769004549%3B1769609349%26q-key-time%3D1769004549%3B1769609349%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3Dc090b0abca8233d258873153062f7d6cb52dd087&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "和风热血漫",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/hfrxm.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1769004549%3B1769609349%26q-key-time%3D1769004549%3B1769609349%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D648c0d844c79bf915e70bf2a51ade46d1aee2c65&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "硬核美漫风",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/yhmmf.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1769004549%3B1769609349%26q-key-time%3D1769004549%3B1769609349%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3Da2c3a63d1dd1db13e189ae6cd4a7b2564c26703e&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "潮酷水墨漫",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/cksmm.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1769004549%3B1769609349%26q-key-time%3D1769004549%3B1769609349%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D21d70a39a67f98efc353f4bad680268368cc81fd&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "暗黑战斗风",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/ahzdf.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1769004549%3B1769609349%26q-key-time%3D1769004549%3B1769609349%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D1d563b95588fa3a9a63e5ab0bd5bd2d9f41628ad&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "赛博故障风",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/sbgzf.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1769004549%3B1769609349%26q-key-time%3D1769004549%3B1769609349%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D84973a6a75fbf592248133429454e52804041964&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "精品韩漫风",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/jphmf.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1769004549%3B1769609349%26q-key-time%3D1769004549%3B1769609349%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3Dda4f0124a198bb0e6018dfefe739aa0ecc884073&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "史诗CG写实",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/sscgxs.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1769004549%3B1769609349%26q-key-time%3D1769004549%3B1769609349%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D733144de79c12cf71e0ca3eb071687cf0ec5cbd4&imageMogr2%2Fformat%2Fwebp",
		},
		{
			Name:     "现实摄影",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/xssy.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1769004549%3B1769609349%26q-key-time%3D1769004549%3B1769609349%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3Dfd2ea1311707a8b5ea0647e8acc89b44a0f29d96&imageMogr2%2Fformat%2Fwebp",
		},
	}
}
