package services

type StyleSeed struct {
	Name     string
	ImageURL string
	PromptZh string
	PromptEn string
}

func DefaultStyleSeeds() []StyleSeed {
	return []StyleSeed{
		{
			Name:     "日漫异世界风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/ri_man_yi_shi_jie.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768984099%3B1769588899%26q-key-time%3D1768984099%3B1769588899%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3De32d9723d7e412124c1bcfb0690f2cd4516f319c&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "日漫异世界风格，二次元插画，清晰线稿，赛璐璐上色，奇幻世界观",
			PromptEn: "anime isekai illustration, clean line art, cel shading, fantasy world, vibrant colors",
		},
		{
			Name:     "奇幻卡通风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/qi_huan_ka_tong.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768984099%3B1769588899%26q-key-time%3D1768984099%3B1769588899%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D7cf50f12a87fdeedb10cbf5e091cf854ba268864&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "奇幻卡通风格，卡通插画，夸张造型，明亮色彩，童话氛围",
			PromptEn: "fantasy cartoon illustration, stylized shapes, bright colors, fairy-tale mood",
		},
		{
			Name:     "国漫仙侠风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/guo_man_xian_xia.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768984099%3B1769588899%26q-key-time%3D1768984099%3B1769588899%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3Db8a7740cea05375ff8cc62527eaa9afdeed23ecf&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "国漫仙侠风格，东方奇幻，飘逸线条，古风服饰，灵气氛围",
			PromptEn: "Chinese xianxia animation style, oriental fantasy, flowing lines, traditional costumes, ethereal mood",
		},
		{
			Name:     "CG史诗风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/cgwrzs.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D68a9b8771a4215c98cda359e0b2b7882935f665b&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "CG史诗风格，高精度CG渲染，电影级光影，宏大场景",
			PromptEn: "epic CGI style, high-detail rendering, cinematic lighting, grand scale",
		},
		{
			Name:     "极细线条韩漫",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/ylqts.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D0789b3738d2d23cc092ebc551859db0e48942f69&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "极细线条韩漫风格，细线稿，柔和配色，干净画面",
			PromptEn: "Korean webtoon thin line art, delicate outlines, soft palette, clean composition",
		},
		{
			Name:     "韩漫古装风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/zqfm.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3Dbcf4269cfd71152ab20e1f55f32f8bd9cb2dcff3&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "韩漫古装风格，细腻线稿，柔和渲染，古装人物与宫廷氛围",
			PromptEn: "Korean webtoon historical style, refined line art, soft shading, period costumes",
		},
		{
			Name:     "韩漫都市风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/hmds.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D40ede0c347c2cd64c2e59d54a7efbe41782f4a62&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "韩漫都市风格，现代都市题材，细腻线稿，柔和渲染",
			PromptEn: "Korean webtoon urban style, modern city theme, refined line art, soft shading",
		},
		{
			Name:     "动漫通用风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/dmty.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3Df9abebb9879223311fe67a7e0f1c3ee3a4810b80&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "动漫通用风格，二次元插画，清晰线稿，赛璐璐上色",
			PromptEn: "anime illustration, clean line art, cel shading, vivid colors",
		},
		{
			Name:     "古风水墨风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/gfsm.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D06c5d76995bb5f9d00b9108a2f4d22b13c158140&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "古风水墨风格，水墨晕染，留白构图，宣纸质感",
			PromptEn: "Chinese ink wash style, brush strokes, soft gradients, elegant negative space",
		},
		{
			Name:     "国风卡通风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/guo_man_ka_tong.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768984099%3B1769588899%26q-key-time%3D1768984099%3B1769588899%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3Dec613015ef9bf6f052fda8473f20e6ac0f8bfabe&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "国风卡通风格，东方元素卡通化，明快色彩，简洁线条",
			PromptEn: "Chinese-style cartoon, stylized oriental motifs, bright colors, clean lines",
		},
		{
			Name:     "都市动漫风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/wdeq.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D322ce2e2fae9bc634d85dea0996e9f02232da1f7&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "都市动漫风格，现代城市题材，二次元插画，清晰线稿",
			PromptEn: "urban anime style, modern city theme, clean line art, cel shading",
		},
		{
			Name:     "校园卡通风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/npxy.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3Def2f6451e05b656095811346572f18ec9f2996d1&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "校园卡通风格，青春校园题材，卡通插画，明快配色",
			PromptEn: "campus cartoon style, youthful theme, bright palette, stylized illustration",
		},
		{
			Name:     "CG都市风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/cgds.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D910919da27bb69aad72e2b70dab5d42cd39b2418&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "CG都市风格，高精度CG渲染，现代都市环境，电影感光影",
			PromptEn: "urban CGI style, high-detail rendering, modern city environment, cinematic lighting",
		},
		{
			Name:     "美漫风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/mm.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D177d8f4ce94cc5333052022d4345facbb37409ec&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "美漫风格，粗线条，强对比阴影，漫画网点质感",
			PromptEn: "American comic style, bold inks, high-contrast shadows, halftone texture",
		},
		{
			Name:     "2d平面风格",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/2dpm.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1768558282%3B1769163082%26q-key-time%3D1768558282%3B1769163082%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D18757cc57fa900dfe60ca5a1c03f55bacea767eb&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "2D平面风格，扁平化插画，简洁形状，干净配色",
			PromptEn: "2D flat illustration, simple shapes, clean colors, minimal shading",
		},
		{
			Name:     "唯美光影二次元",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/wmgyecy.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1769004549%3B1769609349%26q-key-time%3D1769004549%3B1769609349%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3Dc090b0abca8233d258873153062f7d6cb52dd087&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "唯美光影二次元，柔和光晕，细腻渐变，梦幻氛围",
			PromptEn: "beautiful anime lighting, soft glow, delicate gradients, dreamy atmosphere",
		},
		{
			Name:     "和风热血漫",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/hfrxm.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1769004549%3B1769609349%26q-key-time%3D1769004549%3B1769609349%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D648c0d844c79bf915e70bf2a51ade46d1aee2c65&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "和风热血漫，日漫少年热血风，强烈动感，夸张表情",
			PromptEn: "Japanese shonen anime style, dynamic action, bold expressions, energetic mood",
		},
		{
			Name:     "硬核美漫风",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/yhmmf.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1769004549%3B1769609349%26q-key-time%3D1769004549%3B1769609349%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3Da2c3a63d1dd1db13e189ae6cd4a7b2564c26703e&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "硬核美漫风，深色调，硬朗线条，粗粝质感，力量感",
			PromptEn: "gritty American comic style, dark palette, rugged textures, strong contrast",
		},
		{
			Name:     "潮酷水墨漫",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/cksmm.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1769004549%3B1769609349%26q-key-time%3D1769004549%3B1769609349%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D21d70a39a67f98efc353f4bad680268368cc81fd&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "潮酷水墨漫，水墨与潮流融合，泼墨纹理，酷感配色",
			PromptEn: "modern ink wash comic, ink splashes, trendy palette, bold contrast",
		},
		{
			Name:     "暗黑战斗风",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/ahzdf.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1769004549%3B1769609349%26q-key-time%3D1769004549%3B1769609349%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D1d563b95588fa3a9a63e5ab0bd5bd2d9f41628ad&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "暗黑战斗风，低饱和暗色调，强对比光影，压迫感",
			PromptEn: "dark battle style, low saturation, high contrast lighting, intense mood",
		},
		{
			Name:     "赛博故障风",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/sbgzf.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1769004549%3B1769609349%26q-key-time%3D1769004549%3B1769609349%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D84973a6a75fbf592248133429454e52804041964&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "赛博故障风，霓虹色调，电子噪点，故障特效，未来科技感",
			PromptEn: "cyberpunk glitch style, neon palette, digital noise, glitch effects",
		},
		{
			Name:     "精品韩漫风",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/jphmf.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1769004549%3B1769609349%26q-key-time%3D1769004549%3B1769609349%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3Dda4f0124a198bb0e6018dfefe739aa0ecc884073&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "精品韩漫风，精致线稿，柔和光影，细腻质感",
			PromptEn: "premium Korean webtoon style, refined line art, soft lighting, polished finish",
		},
		{
			Name:     "史诗CG写实",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/sscgxs.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1769004549%3B1769609349%26q-key-time%3D1769004549%3B1769609349%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3D733144de79c12cf71e0ca3eb071687cf0ec5cbd4&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "史诗CG写实，高写实CG渲染，电影级光影，宏大场景",
			PromptEn: "epic realistic CGI, cinematic lighting, high-detail realism, grand scale",
		},
		{
			Name:     "现实摄影",
			ImageURL: "https://huimeng-1351980869.cos.ap-beijing.myqcloud.com/script_task/system/style_lora/xssy.png?sign=q-sign-algorithm%3Dsha1%26q-ak%3DAKIDU0JQK0Sh4Wq57vZ0DCxXv51nLgmfNtBG%26q-sign-time%3D1769004549%3B1769609349%26q-key-time%3D1769004549%3B1769609349%26q-header-list%3Dhost%26q-url-param-list%3Dimagemogr2%252fformat%252fwebp%26q-signature%3Dfd2ea1311707a8b5ea0647e8acc89b44a0f29d96&imageMogr2%2Fformat%2Fwebp",
			PromptZh: "现实摄影，商业棚拍，写实人像，真实光影与材质",
			PromptEn: "photorealistic photography, studio lighting, realistic portrait, natural skin texture",
		},
	}
}
