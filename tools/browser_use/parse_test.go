package browser_use

import "testing"

func TestPreCleanHTML(t *testing.T) {
	html := `<!doctype html>
<html lang=zh class=no-js data-theme-init>
    <head>
        <meta charset=utf-8>
        <meta name=viewport content="width=device-width,initial-scale=1,shrink-to-fit=no">
        <meta name=generator content="Hugo 0.151.0">
        <script>
            (function() {
                const t = "td-color-theme"
                  , n = localStorage.getItem(t);
                let e = n || (window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light");
                e === "auto" && (e = window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light"),
                document.documentElement.setAttribute("data-bs-theme", e)
            }
            )()
        </script>
        <meta name=color-scheme content="light dark">
        <style>
            html {
                background: Canvas;
                color: CanvasText
            }

            @media(prefers-color-scheme: dark) {
                html {
                    background:#0b0d12;
                    color: #e6e6e6
                }
            }

            html[data-theme-init] * {
                transition: none!important
            }
        </style>
        <meta name=ROBOTS content="INDEX, FOLLOW">
        <link rel="shortcut icon" href=/favicons/favicon.ico>
        <link rel=apple-touch-icon href=/favicons/apple-touch-icon-180x180.png sizes=180x180>
        <link rel=icon type=image/png href=/favicons/favicon-16x16.png sizes=16x16>
        <link rel=icon type=image/png href=/favicons/favicon-32x32.png sizes=32x32>
        <link rel=icon type=image/png href=/favicons/android-36x36.png sizes=36x36>
        <link rel=icon type=image/png href=/favicons/android-48x48.png sizes=48x48>
        <link rel=icon type=image/png href=/favicons/android-72x72.png sizes=72x72>
        <link rel=icon type=image/png href=/favicons/android-96x96.png sizes=96x96>
        <link rel=icon type=image/png href=/favicons/android-144x144.png sizes=144x144>
        <link rel=icon type=image/png href=/favicons/android-192x192.png sizes=192x192>
        <title>第五章：Middleware（中间件模式） | CloudWeGo</title>
        <meta name=description content="A leading practice for building enterprise cloud native middleware!">
        <meta property="og:url" content="https://www.cloudwego.io/zh/docs/eino/quick_start/chapter_05_middleware/">
        <meta property="og:site_name" content="CloudWeGo">
        <meta property="og:title" content="第五章：Middleware（中间件模式）">
        <meta property="og:description" content='本章目标：理解 Middleware 模式，实现 Tool 错误处理和 ChatModel 重试机制。
为什么需要 Middleware 第四章我们为 Agent 添加了 Tool 能力，让 Agent 能够访问文件系统。但在实际应用场景中，Tool 报错或 ChatModel 报错是常见的现象，例如：
Tool 报错：文件不存在、参数错误、权限不足等 ChatModel 报错：API 限流（429）、网络超时、服务不可用等 问题一：Tool 错误会中断整个流程 当 Tool 执行失败时，错误会直接传播到 Agent，导致整个对话中断：
[tool call] read_file(file_path: "nonexistent.txt") Error: open nonexistent.txt: no such file or directory // 对话中断，用户需要重新开始 问题二：模型调用可能因限流失败 当模型 API 返回 429（Too Many Requests）错误时，整个对话也会中断：
Error: rate limit exceeded (429) // 对话中断 期望的行为 这些报错信息往往不希望直接终止 Agent 流程，而是希望把报错信息给到模型，由模型自动纠错进行下一轮。例如：
[tool call] read_file(file_path: "nonexistent.txt") [tool result] [tool error] open nonexistent.txt: no such file or directory [assistant] 抱歉，文件不存在。让我先列出当前目录的文件... [tool call] glob(pattern: "*") Middleware 的定位 Middleware 模式可以扩展 Tool 和 ChatModel 的行为，非常适合解决这个问题：'>
        <meta property="og:locale" content="zh">
        <meta property="og:type" content="article">
        <meta property="article:section" content="docs">
        <meta property="article:published_time" content="2026-03-16T00:00:00+00:00">
        <meta property="article:modified_time" content="2026-03-16T20:55:36+08:00">
        <meta itemprop=name content="第五章：Middleware（中间件模式）">
        <meta itemprop=description content='本章目标：理解 Middleware 模式，实现 Tool 错误处理和 ChatModel 重试机制。
为什么需要 Middleware 第四章我们为 Agent 添加了 Tool 能力，让 Agent 能够访问文件系统。但在实际应用场景中，Tool 报错或 ChatModel 报错是常见的现象，例如：
Tool 报错：文件不存在、参数错误、权限不足等 ChatModel 报错：API 限流（429）、网络超时、服务不可用等 问题一：Tool 错误会中断整个流程 当 Tool 执行失败时，错误会直接传播到 Agent，导致整个对话中断：
[tool call] read_file(file_path: "nonexistent.txt") Error: open nonexistent.txt: no such file or directory // 对话中断，用户需要重新开始 问题二：模型调用可能因限流失败 当模型 API 返回 429（Too Many Requests）错误时，整个对话也会中断：
Error: rate limit exceeded (429) // 对话中断 期望的行为 这些报错信息往往不希望直接终止 Agent 流程，而是希望把报错信息给到模型，由模型自动纠错进行下一轮。例如：
[tool call] read_file(file_path: "nonexistent.txt") [tool result] [tool error] open nonexistent.txt: no such file or directory [assistant] 抱歉，文件不存在。让我先列出当前目录的文件... [tool call] glob(pattern: "*") Middleware 的定位 Middleware 模式可以扩展 Tool 和 ChatModel 的行为，非常适合解决这个问题：'>
        <meta itemprop=datePublished content="2026-03-16T00:00:00+00:00">
        <meta itemprop=dateModified content="2026-03-16T20:55:36+08:00">
        <meta itemprop=wordCount content="1118">
        <meta name=twitter:card content="summary">
        <meta name=twitter:title content="第五章：Middleware（中间件模式）">
        <meta name=twitter:description content='本章目标：理解 Middleware 模式，实现 Tool 错误处理和 ChatModel 重试机制。
为什么需要 Middleware 第四章我们为 Agent 添加了 Tool 能力，让 Agent 能够访问文件系统。但在实际应用场景中，Tool 报错或 ChatModel 报错是常见的现象，例如：
Tool 报错：文件不存在、参数错误、权限不足等 ChatModel 报错：API 限流（429）、网络超时、服务不可用等 问题一：Tool 错误会中断整个流程 当 Tool 执行失败时，错误会直接传播到 Agent，导致整个对话中断：
[tool call] read_file(file_path: "nonexistent.txt") Error: open nonexistent.txt: no such file or directory // 对话中断，用户需要重新开始 问题二：模型调用可能因限流失败 当模型 API 返回 429（Too Many Requests）错误时，整个对话也会中断：
Error: rate limit exceeded (429) // 对话中断 期望的行为 这些报错信息往往不希望直接终止 Agent 流程，而是希望把报错信息给到模型，由模型自动纠错进行下一轮。例如：
[tool call] read_file(file_path: "nonexistent.txt") [tool result] [tool error] open nonexistent.txt: no such file or directory [assistant] 抱歉，文件不存在。让我先列出当前目录的文件... [tool call] glob(pattern: "*") Middleware 的定位 Middleware 模式可以扩展 Tool 和 ChatModel 的行为，非常适合解决这个问题：'>
        <script async src="https://www.googletagmanager.com/gtag/js?id=G-QYWRQRLPRM"></script>
        <script>
            window.dataLayer = window.dataLayer || [];
            function gtag() {
                dataLayer.push(arguments)
            }
            gtag("js", new Date),
            gtag("config", "G-QYWRQRLPRM")
        </script>
        <script>
            var _hmt = _hmt || [];
            (function() {
                var e, t = document.createElement("script");
                t.src = "https://hm.baidu.com/hm.js?f1808c42af827f368aa7eca3baae6d55",
                e = document.getElementsByTagName("script")[0],
                e.parentNode.insertBefore(t, e)
            }
            )()
        </script>
        <link rel=preload href=/scss/main.min.2ad5ca42a3e8c8d3be18c26c471e45a60f2e90839b3fd2c18757e2622f17b628.css as=style>
        <link href=/scss/main.min.2ad5ca42a3e8c8d3be18c26c471e45a60f2e90839b3fd2c18757e2622f17b628.css rel=stylesheet integrity>
        <script src=/js/jquery.min.js></script>
        <link rel=stylesheet href=/css/docsearch.css>
    </head>
    <body class=td-page>
        <header>
            <nav class="js-navbar-scroll navbar navbar-expand-xl navbar-dark td-navbar">
                <a class=navbar-brand href=/zh/>
                <span class=navbar-logo>
                    <img src=/img/logo.png>
                </span>
</a>
<button class=navbar-toggler type=button data-bs-toggle=collapse data-bs-target=#main_navbar aria-controls=main_navbar aria-expanded=false aria-label="Toggle navigation">
    <span class=navbar-toggler-icon></span>
</button>
<div class="collapse navbar-collapse td-navbar-nav-scroll ms-md-auto" id=main_navbar>
    <ul class="navbar-nav mt-2 mt-lg-0 ms-auto">
        <li class="dropdown sub-menu active">
            <a class="nav-link dropdown-toggle" href=# id=navbarDropdown role=button data-bs-toggle=dropdown aria-haspopup=true aria-expanded=false>
                <span>文档</span>
            </a>
            <div class=dropdown-menu aria-labelledby=navbarDropdown>
                <a class=dropdown-item href=/zh/docs/kitex/>Kitex</a>
                <a class=dropdown-item href=/zh/docs/hertz/>Hertz</a>
                <a class=dropdown-item href=/zh/docs/volo/>Volo</a>
                <a class=dropdown-item href=/zh/docs/eino/>Eino</a>
            </div>
        </li>
        <li class="nav-item me-4 mb-2 mb-lg-0">
            <a class=nav-link href=/zh/about/>
            <span>关于</span>
</a></li>
<li class="nav-item me-4 mb-2 mb-lg-0">
    <a class=nav-link href=/zh/blog/>
    <span>博客</span>
</a></li>
<li class="dropdown sub-menu">
    <a class="nav-link dropdown-toggle" href=# id=navbarDropdown role=button data-bs-toggle=dropdown aria-haspopup=true aria-expanded=false>
        <span>社区</span>
    </a>
    <div class=dropdown-menu aria-labelledby=navbarDropdown>
        <a class=dropdown-item href=/zh/community/overview/>概述</a>
        <a class=dropdown-item href=/zh/community/meeting_notes/>会议记录</a>
        <a class=dropdown-item href=/zh/community/weekly_report/>周报</a>
        <a class=dropdown-item href=/zh/community/past_activities/>往期活动</a>
    </div>
</li>
<li class="nav-item me-4 mb-2 mb-lg-0">
    <a class=nav-link href=/zh/cooperation/>
    <span>用户案例</span>
</a></li>
<li class="dropdown sub-menu">
    <a class="nav-link dropdown-toggle" href=# id=navbarDropdown role=button data-bs-toggle=dropdown aria-haspopup=true aria-expanded=false>
        <span>企业级解决方案</span>
    </a>
    <div class=dropdown-menu aria-labelledby=navbarDropdown>
        <a class=dropdown-item href=https://www.volcengine.com/docs/6431/1469323>可观测解决方案</a>
    </div>
</li>
<li class="dropdown sub-menu">
    <a class="nav-link dropdown-toggle" href=# id=navbarDropdown role=button data-bs-toggle=dropdown aria-haspopup=true aria-expanded=false>
        <span>安全</span>
    </a>
    <div class=dropdown-menu aria-labelledby=navbarDropdown>
        <a class=dropdown-item href=/zh/security/safety-bulletin/>安全公告</a>
        <a class=dropdown-item href=/zh/security/vulnerability-reporting/>漏洞管理</a>
    </div>
</li>
<li class="nav-item dropdown me-2 me-lg-4">
    <a class="nav-link dropdown-toggle d-flex align-items-center" href=# id=navbarLangDropdownLink role=button data-bs-toggle=dropdown aria-haspopup=true aria-expanded=false>
        <span>中文</span>
    </a>
    <div class="dropdown-menu dropdown-menu-end" aria-labelledby=navbarLangDropdownLink>
        <a class=dropdown-item href=/docs/eino/quick_start/chapter_05_middleware/>English</a>
    </div>
</li>
<li class="nav-item dropdown td-navbar__light-dark-menu">
    <svg class="d-none">
        <symbol id="check2" viewBox="0 0 16 16">
            <path d="M13.854 3.646a.5.5.0 010 .708l-7 7a.5.5.0 01-.708.0l-3.5-3.5a.5.5.0 11.708-.708L6.5 10.293l6.646-6.647a.5.5.0 01.708.0z"/>
        </symbol>
        <symbol id="circle-half" viewBox="0 0 16 16">
            <path d="M8 15A7 7 0 108 1v14zm0 1A8 8 0 118 0a8 8 0 010 16z"/>
        </symbol>
        <symbol id="moon-stars-fill" viewBox="0 0 16 16">
            <path d="M6 .278a.768.768.0 01.08.858 7.208 7.208.0 00-.878 3.46c0 4.021 3.278 7.277 7.318 7.277.527.0 1.04-.055 1.533-.16a.787.787.0 01.81.316.733.733.0 01-.031.893A8.349 8.349.0 018.344 16C3.734 16 0 12.286.0 7.71.0 4.266 2.114 1.312 5.124.06A.752.752.0 016 .278z"/>
            <path d="M10.794 3.148a.217.217.0 01.412.0l.387 1.162c.173.518.579.924 1.097 1.097l1.162.387a.217.217.0 010 .412l-1.162.387A1.734 1.734.0 0011.593 7.69l-.387 1.162a.217.217.0 01-.412.0l-.387-1.162A1.734 1.734.0 009.31 6.593l-1.162-.387a.217.217.0 010-.412l1.162-.387a1.734 1.734.0 001.097-1.097l.387-1.162zM13.863.099a.145.145.0 01.274.0l.258.774c.115.346.386.617.732.732l.774.258a.145.145.0 010 .274l-.774.258a1.156 1.156.0 00-.732.732l-.258.774a.145.145.0 01-.274.0l-.258-.774a1.156 1.156.0 00-.732-.732l-.774-.258a.145.145.0 010-.274l.774-.258c.346-.115.617-.386.732-.732L13.863.1z"/>
        </symbol>
        <symbol id="sun-fill" viewBox="0 0 16 16">
            <path d="M8 12a4 4 0 100-8 4 4 0 000 8zM8 0a.5.5.0 01.5.5v2a.5.5.0 01-1 0v-2A.5.5.0 018 0zm0 13a.5.5.0 01.5.5v2a.5.5.0 01-1 0v-2A.5.5.0 018 13zm8-5a.5.5.0 01-.5.5h-2a.5.5.0 010-1h2a.5.5.0 01.5.5zM3 8a.5.5.0 01-.5.5h-2a.5.5.0 010-1h2A.5.5.0 013 8zm10.657-5.657a.5.5.0 010 .707l-1.414 1.415a.5.5.0 11-.707-.708l1.414-1.414a.5.5.0 01.707.0zm-9.193 9.193a.5.5.0 010 .707L3.05 13.657a.5.5.0 01-.707-.707l1.414-1.414a.5.5.0 01.707.0zm9.193 2.121a.5.5.0 01-.707.0l-1.414-1.414a.5.5.0 01.707-.707l1.414 1.414a.5.5.0 010 .707zM4.464 4.465a.5.5.0 01-.707.0L2.343 3.05a.5.5.0 11.707-.707l1.414 1.414a.5.5.0 010 .708z"/>
        </symbol>
    </svg>
    <button class="btn btn-link nav-link dropdown-toggle d-flex align-items-center" id=bd-theme type=button aria-expanded=false data-bs-toggle=dropdown data-bs-offset=0,0 aria-label="Toggle theme (auto)">
        <svg class="bi my-1 theme-icon-active">
            <use href="#circle-half"/>
        </svg>
    </button>
    <ul class="dropdown-menu dropdown-menu-end" aria-labelledby=bd-theme>
        <li>
            <button type=button class="dropdown-item d-flex align-items-center" data-bs-theme-value=light aria-pressed=false>
                <svg class="bi me-2 opacity-50">
                    <use href="#sun-fill"/>
                </svg>
                Light

                <svg class="bi ms-auto d-none">
                    <use href="#check2"/>
                </svg>
            </button>
        </li>
        <li>
            <button type=button class="dropdown-item d-flex align-items-center" data-bs-theme-value=dark aria-pressed=false>
                <svg class="bi me-2 opacity-50">
                    <use href="#moon-stars-fill"/>
                </svg>
                Dark

                <svg class="bi ms-auto d-none">
                    <use href="#check2"/>
                </svg>
            </button>
        </li>
        <li>
            <button type=button class="dropdown-item d-flex align-items-center active" data-bs-theme-value=auto aria-pressed=true>
                <svg class="bi me-2 opacity-50">
                    <use href="#circle-half"/>
                </svg>
                Auto

                <svg class="bi ms-auto d-none">
                    <use href="#check2"/>
                </svg>
            </button>
        </li>
    </ul>
</li>
</ul></div>
<div class="navbar-nav d-none d-lg-block">
    <div id=docsearch></div>
</div>
</nav></header>
<div class="container-fluid td-outer">
    <div class=td-main>
        <div class="row flex-xl-nowrap">
            <aside class="col-12 col-md-3 col-xl-2 td-sidebar d-print-none">
                <div id=td-sidebar-menu class=td-sidebar__inner>
                    <div id=content-mobile>
                        <form class="td-sidebar__search d-flex align-items-center">
                            <div id=docsearch></div>
                            <button class="btn btn-link td-sidebar__toggle d-md-none p-0 ms-3 fas fa-bars" type=button data-bs-toggle=collapse data-bs-target=#td-section-nav aria-controls=td-docs-nav aria-expanded=false aria-label="Toggle section navigation"></button>
                        </form>
                    </div>
                    <div id=content-desktop></div>
                    <nav class="collapse td-sidebar-nav foldable-nav" id=td-section-nav>
                        <div class="nav-item dropdown d-block d-lg-none px-3 mb-3">
                            <a class="nav-link dropdown-toggle d-flex align-items-center" href=# id=navbarLangDropdownLink role=button data-bs-toggle=dropdown aria-haspopup=true aria-expanded=false>
                                <span>中文</span>
                            </a>
                            <div class="dropdown-menu dropdown-menu-start" aria-labelledby=navbarLangDropdownLink>
                                <a class=dropdown-item href=/docs/eino/quick_start/chapter_05_middleware/>English</a>
                            </div>
                        </div>
                        <ul class="td-sidebar-nav__section pe-md-3 ul-0">
                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section with-child active-path" id=m-zhdocs-li>
                                <a href=/zh/docs/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section tree-root" id=m-zhdocs>
                                    <span>文档</span>
                                </a>
                                <ul class=ul-1>
                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section with-child active-path" id=m-zhdocseino-li>
                                        <input type=checkbox id=m-zhdocseino-check checked>
                                        <label for=m-zhdocseino-check>
                                            <a href=/zh/docs/eino/ title="Eino 用户手册" class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseino>
                                                <span>Eino</span>
                                            </a>
                                        </label>
                                        <ul class="ul-2 foldable">
                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section with-child" id=m-zhdocseinooverview-li>
                                                <input type=checkbox id=m-zhdocseinooverview-check checked>
                                                <label for=m-zhdocseinooverview-check>
                                                    <a href=/zh/docs/eino/overview/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinooverview>
                                                        <span>概述</span>
                                                    </a>
                                                </label>
                                                <ul class="ul-3 foldable">
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinooverviewbytedance_eino_practice-li>
                                                        <input type=checkbox id=m-zhdocseinooverviewbytedance_eino_practice-check>
                                                        <label for=m-zhdocseinooverviewbytedance_eino_practice-check>
                                                            <a href=/zh/docs/eino/overview/bytedance_eino_practice/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinooverviewbytedance_eino_practice>
                                                                <span>字节跳动大模型应用 Go 开发框架 —— Eino 实践</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoovervieweino_open_source-li>
                                                        <input type=checkbox id=m-zhdocseinoovervieweino_open_source-check>
                                                        <label for=m-zhdocseinoovervieweino_open_source-check>
                                                            <a href=/zh/docs/eino/overview/eino_open_source/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinoovervieweino_open_source>
                                                                <span>大语言模型应用开发框架 —— Eino 正式开源！</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoovervieweino_adk0_1-li>
                                                        <input type=checkbox id=m-zhdocseinoovervieweino_adk0_1-check>
                                                        <label for=m-zhdocseinoovervieweino_adk0_1-check>
                                                            <a href=/zh/docs/eino/overview/eino_adk0_1/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinoovervieweino_adk0_1>
                                                                <span>Eino ADK：一文搞定 AI Agent 核心设计模式，从 0 到 1 搭建智能体系统</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoovervieweino_adk_excel_agent-li>
                                                        <input type=checkbox id=m-zhdocseinoovervieweino_adk_excel_agent-check>
                                                        <label for=m-zhdocseinoovervieweino_adk_excel_agent-check>
                                                            <a href=/zh/docs/eino/overview/eino_adk_excel_agent/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinoovervieweino_adk_excel_agent>
                                                                <span>用 Eino ADK 构建你的第一个 AI 智能体：从 Excel Agent 实战开始</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinooverviewgraph_or_agent-li>
                                                        <input type=checkbox id=m-zhdocseinooverviewgraph_or_agent-check>
                                                        <label for=m-zhdocseinooverviewgraph_or_agent-check>
                                                            <a href=/zh/docs/eino/overview/graph_or_agent/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinooverviewgraph_or_agent>
                                                                <span>Agent 还是 Graph？AI 应用路线辨析</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                </ul>
                                            </li>
                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section with-child active-path" id=m-zhdocseinoquick_start-li>
                                                <input type=checkbox id=m-zhdocseinoquick_start-check checked>
                                                <label for=m-zhdocseinoquick_start-check>
                                                    <a href=/zh/docs/eino/quick_start/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinoquick_start>
                                                        <span>快速开始</span>
                                                    </a>
                                                </label>
                                                <ul class="ul-3 foldable">
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoquick_startchapter_01_chatmodel_and_message-li>
                                                        <input type=checkbox id=m-zhdocseinoquick_startchapter_01_chatmodel_and_message-check>
                                                        <label for=m-zhdocseinoquick_startchapter_01_chatmodel_and_message-check>
                                                            <a href=/zh/docs/eino/quick_start/chapter_01_chatmodel_and_message/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinoquick_startchapter_01_chatmodel_and_message>
                                                                <span>第一章：ChatModel 与 Message（Console）</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoquick_startchapter_02_chatmodelagent_runner_agentevent-li>
                                                        <input type=checkbox id=m-zhdocseinoquick_startchapter_02_chatmodelagent_runner_agentevent-check>
                                                        <label for=m-zhdocseinoquick_startchapter_02_chatmodelagent_runner_agentevent-check>
                                                            <a href=/zh/docs/eino/quick_start/chapter_02_chatmodelagent_runner_agentevent/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinoquick_startchapter_02_chatmodelagent_runner_agentevent>
                                                                <span>第二章：ChatModelAgent、Runner、AgentEvent（Console 多轮）</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoquick_startchapter_03_memory_and_session-li>
                                                        <input type=checkbox id=m-zhdocseinoquick_startchapter_03_memory_and_session-check>
                                                        <label for=m-zhdocseinoquick_startchapter_03_memory_and_session-check>
                                                            <a href=/zh/docs/eino/quick_start/chapter_03_memory_and_session/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinoquick_startchapter_03_memory_and_session>
                                                                <span>第三章：Memory 与 Session（持久化对话）</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoquick_startchapter_04_tool_and_filesystem-li>
                                                        <input type=checkbox id=m-zhdocseinoquick_startchapter_04_tool_and_filesystem-check>
                                                        <label for=m-zhdocseinoquick_startchapter_04_tool_and_filesystem-check>
                                                            <a href=/zh/docs/eino/quick_start/chapter_04_tool_and_filesystem/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinoquick_startchapter_04_tool_and_filesystem>
                                                                <span>第四章：Tool 与文件系统访问</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child active-path" id=m-zhdocseinoquick_startchapter_05_middleware-li>
                                                        <input type=checkbox id=m-zhdocseinoquick_startchapter_05_middleware-check checked>
                                                        <label for=m-zhdocseinoquick_startchapter_05_middleware-check>
                                                            <a href=/zh/docs/eino/quick_start/chapter_05_middleware/ class="align-left ps-0 active td-sidebar-link td-sidebar-link__page" id=m-zhdocseinoquick_startchapter_05_middleware>
                                                                <span class=td-sidebar-nav-active-item>第五章：Middleware（中间件模式）</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoquick_startchapter_06_callback_and_trace-li>
                                                        <input type=checkbox id=m-zhdocseinoquick_startchapter_06_callback_and_trace-check>
                                                        <label for=m-zhdocseinoquick_startchapter_06_callback_and_trace-check>
                                                            <a href=/zh/docs/eino/quick_start/chapter_06_callback_and_trace/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinoquick_startchapter_06_callback_and_trace>
                                                                <span>第六章：Callback 与 Trace（可观测性）</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoquick_startchapter_07_interrupt_resume-li>
                                                        <input type=checkbox id=m-zhdocseinoquick_startchapter_07_interrupt_resume-check>
                                                        <label for=m-zhdocseinoquick_startchapter_07_interrupt_resume-check>
                                                            <a href=/zh/docs/eino/quick_start/chapter_07_interrupt_resume/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinoquick_startchapter_07_interrupt_resume>
                                                                <span>第七章：Interrupt/Resume（中断与恢复）</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoquick_startchapter_08_graph_tool-li>
                                                        <input type=checkbox id=m-zhdocseinoquick_startchapter_08_graph_tool-check>
                                                        <label for=m-zhdocseinoquick_startchapter_08_graph_tool-check>
                                                            <a href=/zh/docs/eino/quick_start/chapter_08_graph_tool/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinoquick_startchapter_08_graph_tool>
                                                                <span>第八章：Graph Tool（复杂工作流）</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoquick_startchapter_09_skill_console-li>
                                                        <input type=checkbox id=m-zhdocseinoquick_startchapter_09_skill_console-check>
                                                        <label for=m-zhdocseinoquick_startchapter_09_skill_console-check>
                                                            <a href=/zh/docs/eino/quick_start/chapter_09_skill_console/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinoquick_startchapter_09_skill_console>
                                                                <span>第九章：Skill（Console）</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoquick_startchapter_09_a2ui_protocol-li>
                                                        <input type=checkbox id=m-zhdocseinoquick_startchapter_09_a2ui_protocol-check>
                                                        <label for=m-zhdocseinoquick_startchapter_09_a2ui_protocol-check>
                                                            <a href=/zh/docs/eino/quick_start/chapter_09_a2ui_protocol/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinoquick_startchapter_09_a2ui_protocol>
                                                                <span>第十章：A2UI 协议（流式 UI 组件）</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                </ul>
                                            </li>
                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocookbook-li>
                                                <input type=checkbox id=m-zhdocseinocookbook-check checked>
                                                <label for=m-zhdocseinocookbook-check>
                                                    <a href=/zh/docs/eino/cookbook/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocookbook>
                                                        <span>Cookbook</span>
                                                    </a>
                                                </label>
                                            </li>
                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section with-child" id=m-zhdocseinocore_modules-li>
                                                <input type=checkbox id=m-zhdocseinocore_modules-check checked>
                                                <label for=m-zhdocseinocore_modules-check>
                                                    <a href=/zh/docs/eino/core_modules/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinocore_modules>
                                                        <span>核心模块</span>
                                                    </a>
                                                </label>
                                                <ul class="ul-3 foldable">
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section with-child" id=m-zhdocseinocore_modulescomponents-li>
                                                        <input type=checkbox id=m-zhdocseinocore_modulescomponents-check>
                                                        <label for=m-zhdocseinocore_modulescomponents-check>
                                                            <a href=/zh/docs/eino/core_modules/components/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinocore_modulescomponents>
                                                                <span>Components 组件</span>
                                                            </a>
                                                        </label>
                                                        <ul class="ul-4 foldable">
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section with-child" id=m-zhdocseinocore_modulescomponentsdocument_loader_guide-li>
                                                                <input type=checkbox id=m-zhdocseinocore_modulescomponentsdocument_loader_guide-check>
                                                                <label for=m-zhdocseinocore_modulescomponentsdocument_loader_guide-check>
                                                                    <a href=/zh/docs/eino/core_modules/components/document_loader_guide/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinocore_modulescomponentsdocument_loader_guide>
                                                                        <span>Document Loader 使用说明</span>
                                                                    </a>
                                                                </label>
                                                                <ul class="ul-5 foldable">
                                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_modulescomponentsdocument_loader_guidedocument_parser_interface_guide-li>
                                                                        <input type=checkbox id=m-zhdocseinocore_modulescomponentsdocument_loader_guidedocument_parser_interface_guide-check>
                                                                        <label for=m-zhdocseinocore_modulescomponentsdocument_loader_guidedocument_parser_interface_guide-check>
                                                                            <a href=/zh/docs/eino/core_modules/components/document_loader_guide/document_parser_interface_guide/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_modulescomponentsdocument_loader_guidedocument_parser_interface_guide>
                                                                                <span>Document Parser 接口使用说明</span>
                                                                            </a>
                                                                        </label>
                                                                    </li>
                                                                </ul>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_modulescomponentsembedding_guide-li>
                                                                <input type=checkbox id=m-zhdocseinocore_modulescomponentsembedding_guide-check>
                                                                <label for=m-zhdocseinocore_modulescomponentsembedding_guide-check>
                                                                    <a href=/zh/docs/eino/core_modules/components/embedding_guide/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_modulescomponentsembedding_guide>
                                                                        <span>Embedding 使用说明</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_modulescomponentsdocument_transformer_guide-li>
                                                                <input type=checkbox id=m-zhdocseinocore_modulescomponentsdocument_transformer_guide-check>
                                                                <label for=m-zhdocseinocore_modulescomponentsdocument_transformer_guide-check>
                                                                    <a href=/zh/docs/eino/core_modules/components/document_transformer_guide/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_modulescomponentsdocument_transformer_guide>
                                                                        <span>Document Transformer 使用说明</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_modulescomponentslambda_guide-li>
                                                                <input type=checkbox id=m-zhdocseinocore_modulescomponentslambda_guide-check>
                                                                <label for=m-zhdocseinocore_modulescomponentslambda_guide-check>
                                                                    <a href=/zh/docs/eino/core_modules/components/lambda_guide/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_modulescomponentslambda_guide>
                                                                        <span>Lambda 使用说明</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_modulescomponentsindexer_guide-li>
                                                                <input type=checkbox id=m-zhdocseinocore_modulescomponentsindexer_guide-check>
                                                                <label for=m-zhdocseinocore_modulescomponentsindexer_guide-check>
                                                                    <a href=/zh/docs/eino/core_modules/components/indexer_guide/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_modulescomponentsindexer_guide>
                                                                        <span>Indexer 使用说明</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_modulescomponentsretriever_guide-li>
                                                                <input type=checkbox id=m-zhdocseinocore_modulescomponentsretriever_guide-check>
                                                                <label for=m-zhdocseinocore_modulescomponentsretriever_guide-check>
                                                                    <a href=/zh/docs/eino/core_modules/components/retriever_guide/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_modulescomponentsretriever_guide>
                                                                        <span>Retriever 使用说明</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_modulescomponentschat_template_guide-li>
                                                                <input type=checkbox id=m-zhdocseinocore_modulescomponentschat_template_guide-check>
                                                                <label for=m-zhdocseinocore_modulescomponentschat_template_guide-check>
                                                                    <a href=/zh/docs/eino/core_modules/components/chat_template_guide/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_modulescomponentschat_template_guide>
                                                                        <span>ChatTemplate 使用说明</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_modulescomponentschat_model_guide-li>
                                                                <input type=checkbox id=m-zhdocseinocore_modulescomponentschat_model_guide-check>
                                                                <label for=m-zhdocseinocore_modulescomponentschat_model_guide-check>
                                                                    <a href=/zh/docs/eino/core_modules/components/chat_model_guide/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_modulescomponentschat_model_guide>
                                                                        <span>ChatModel 使用说明</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section with-child" id=m-zhdocseinocore_modulescomponentstools_node_guide-li>
                                                                <input type=checkbox id=m-zhdocseinocore_modulescomponentstools_node_guide-check>
                                                                <label for=m-zhdocseinocore_modulescomponentstools_node_guide-check>
                                                                    <a href=/zh/docs/eino/core_modules/components/tools_node_guide/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinocore_modulescomponentstools_node_guide>
                                                                        <span>ToolsNode &amp;Tool 使用说明</span>
                                                                    </a>
                                                                </label>
                                                                <ul class="ul-5 foldable">
                                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_modulescomponentstools_node_guidehow_to_create_a_tool-li>
                                                                        <input type=checkbox id=m-zhdocseinocore_modulescomponentstools_node_guidehow_to_create_a_tool-check>
                                                                        <label for=m-zhdocseinocore_modulescomponentstools_node_guidehow_to_create_a_tool-check>
                                                                            <a href=/zh/docs/eino/core_modules/components/tools_node_guide/how_to_create_a_tool/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_modulescomponentstools_node_guidehow_to_create_a_tool>
                                                                                <span>如何创建一个 tool ?</span>
                                                                            </a>
                                                                        </label>
                                                                    </li>
                                                                </ul>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_modulescomponentsagentic_chat_model_guide-li>
                                                                <input type=checkbox id=m-zhdocseinocore_modulescomponentsagentic_chat_model_guide-check>
                                                                <label for=m-zhdocseinocore_modulescomponentsagentic_chat_model_guide-check>
                                                                    <a href=/zh/docs/eino/core_modules/components/agentic_chat_model_guide/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_modulescomponentsagentic_chat_model_guide>
                                                                        <span>AgenticModel 使用说明[Beta]</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_modulescomponentsagentic_chat_template_guide-li>
                                                                <input type=checkbox id=m-zhdocseinocore_modulescomponentsagentic_chat_template_guide-check>
                                                                <label for=m-zhdocseinocore_modulescomponentsagentic_chat_template_guide-check>
                                                                    <a href=/zh/docs/eino/core_modules/components/agentic_chat_template_guide/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_modulescomponentsagentic_chat_template_guide>
                                                                        <span>AgenticChatTemplate 使用说明[Beta]</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_modulescomponentsagentic_tools_node_guide-li>
                                                                <input type=checkbox id=m-zhdocseinocore_modulescomponentsagentic_tools_node_guide-check>
                                                                <label for=m-zhdocseinocore_modulescomponentsagentic_tools_node_guide-check>
                                                                    <a href=/zh/docs/eino/core_modules/components/agentic_tools_node_guide/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_modulescomponentsagentic_tools_node_guide>
                                                                        <span>AgenticToolsNode &amp;Tool 使用说明[Beta]</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                        </ul>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section with-child" id=m-zhdocseinocore_moduleschain_and_graph_orchestration-li>
                                                        <input type=checkbox id=m-zhdocseinocore_moduleschain_and_graph_orchestration-check>
                                                        <label for=m-zhdocseinocore_moduleschain_and_graph_orchestration-check>
                                                            <a href=/zh/docs/eino/core_modules/chain_and_graph_orchestration/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinocore_moduleschain_and_graph_orchestration>
                                                                <span>Chain & Graph & Workflow 编排功能</span>
                                                            </a>
                                                        </label>
                                                        <ul class="ul-4 foldable">
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleschain_and_graph_orchestrationchain_graph_introduction-li>
                                                                <input type=checkbox id=m-zhdocseinocore_moduleschain_and_graph_orchestrationchain_graph_introduction-check>
                                                                <label for=m-zhdocseinocore_moduleschain_and_graph_orchestrationchain_graph_introduction-check>
                                                                    <a href=/zh/docs/eino/core_modules/chain_and_graph_orchestration/chain_graph_introduction/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleschain_and_graph_orchestrationchain_graph_introduction>
                                                                        <span>Chain/Graph 编排介绍</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleschain_and_graph_orchestrationorchestration_design_principles-li>
                                                                <input type=checkbox id=m-zhdocseinocore_moduleschain_and_graph_orchestrationorchestration_design_principles-check>
                                                                <label for=m-zhdocseinocore_moduleschain_and_graph_orchestrationorchestration_design_principles-check>
                                                                    <a href=/zh/docs/eino/core_modules/chain_and_graph_orchestration/orchestration_design_principles/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleschain_and_graph_orchestrationorchestration_design_principles>
                                                                        <span>编排的设计理念</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleschain_and_graph_orchestrationworkflow_orchestration_framework-li>
                                                                <input type=checkbox id=m-zhdocseinocore_moduleschain_and_graph_orchestrationworkflow_orchestration_framework-check>
                                                                <label for=m-zhdocseinocore_moduleschain_and_graph_orchestrationworkflow_orchestration_framework-check>
                                                                    <a href=/zh/docs/eino/core_modules/chain_and_graph_orchestration/workflow_orchestration_framework/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleschain_and_graph_orchestrationworkflow_orchestration_framework>
                                                                        <span>Workflow 编排框架</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleschain_and_graph_orchestrationstream_programming_essentials-li>
                                                                <input type=checkbox id=m-zhdocseinocore_moduleschain_and_graph_orchestrationstream_programming_essentials-check>
                                                                <label for=m-zhdocseinocore_moduleschain_and_graph_orchestrationstream_programming_essentials-check>
                                                                    <a href=/zh/docs/eino/core_modules/chain_and_graph_orchestration/stream_programming_essentials/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleschain_and_graph_orchestrationstream_programming_essentials>
                                                                        <span>Eino 流式编程要点</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleschain_and_graph_orchestrationcallback_manual-li>
                                                                <input type=checkbox id=m-zhdocseinocore_moduleschain_and_graph_orchestrationcallback_manual-check>
                                                                <label for=m-zhdocseinocore_moduleschain_and_graph_orchestrationcallback_manual-check>
                                                                    <a href=/zh/docs/eino/core_modules/chain_and_graph_orchestration/callback_manual/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleschain_and_graph_orchestrationcallback_manual>
                                                                        <span>Callback 用户手册</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleschain_and_graph_orchestrationcall_option_capabilities-li>
                                                                <input type=checkbox id=m-zhdocseinocore_moduleschain_and_graph_orchestrationcall_option_capabilities-check>
                                                                <label for=m-zhdocseinocore_moduleschain_and_graph_orchestrationcall_option_capabilities-check>
                                                                    <a href=/zh/docs/eino/core_modules/chain_and_graph_orchestration/call_option_capabilities/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleschain_and_graph_orchestrationcall_option_capabilities>
                                                                        <span>CallOption 能力与规范</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleschain_and_graph_orchestrationcheckpoint_interrupt-li>
                                                                <input type=checkbox id=m-zhdocseinocore_moduleschain_and_graph_orchestrationcheckpoint_interrupt-check>
                                                                <label for=m-zhdocseinocore_moduleschain_and_graph_orchestrationcheckpoint_interrupt-check>
                                                                    <a href=/zh/docs/eino/core_modules/chain_and_graph_orchestration/checkpoint_interrupt/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleschain_and_graph_orchestrationcheckpoint_interrupt>
                                                                        <span>Interrupt & CheckPoint使用手册</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                        </ul>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section with-child" id=m-zhdocseinocore_modulesflow_integration_components-li>
                                                        <input type=checkbox id=m-zhdocseinocore_modulesflow_integration_components-check>
                                                        <label for=m-zhdocseinocore_modulesflow_integration_components-check>
                                                            <a href=/zh/docs/eino/core_modules/flow_integration_components/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinocore_modulesflow_integration_components>
                                                                <span>Flow 集成</span>
                                                            </a>
                                                        </label>
                                                        <ul class="ul-4 foldable">
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_modulesflow_integration_componentsreact_agent_manual-li>
                                                                <input type=checkbox id=m-zhdocseinocore_modulesflow_integration_componentsreact_agent_manual-check>
                                                                <label for=m-zhdocseinocore_modulesflow_integration_componentsreact_agent_manual-check>
                                                                    <a href=/zh/docs/eino/core_modules/flow_integration_components/react_agent_manual/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_modulesflow_integration_componentsreact_agent_manual>
                                                                        <span>ReAct Agent 使用手册</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_modulesflow_integration_componentsmulti_agent_hosting-li>
                                                                <input type=checkbox id=m-zhdocseinocore_modulesflow_integration_componentsmulti_agent_hosting-check>
                                                                <label for=m-zhdocseinocore_modulesflow_integration_componentsmulti_agent_hosting-check>
                                                                    <a href=/zh/docs/eino/core_modules/flow_integration_components/multi_agent_hosting/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_modulesflow_integration_componentsmulti_agent_hosting>
                                                                        <span>Host Multi-Agent</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                        </ul>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section with-child" id=m-zhdocseinocore_moduleseino_adk-li>
                                                        <input type=checkbox id=m-zhdocseinocore_moduleseino_adk-check>
                                                        <label for=m-zhdocseinocore_moduleseino_adk-check>
                                                            <a href=/zh/docs/eino/core_modules/eino_adk/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinocore_moduleseino_adk>
                                                                <span>ADK - Agent Development Kit</span>
                                                            </a>
                                                        </label>
                                                        <ul class="ul-4 foldable">
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkagent_quickstart-li>
                                                                <input type=checkbox id=m-zhdocseinocore_moduleseino_adkagent_quickstart-check>
                                                                <label for=m-zhdocseinocore_moduleseino_adkagent_quickstart-check>
                                                                    <a href=/zh/docs/eino/core_modules/eino_adk/agent_quickstart/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkagent_quickstart>
                                                                        <span>Quickstart</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkagent_preview-li>
                                                                <input type=checkbox id=m-zhdocseinocore_moduleseino_adkagent_preview-check>
                                                                <label for=m-zhdocseinocore_moduleseino_adkagent_preview-check>
                                                                    <a href=/zh/docs/eino/core_modules/eino_adk/agent_preview/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkagent_preview>
                                                                        <span>概述</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkagent_interface-li>
                                                                <input type=checkbox id=m-zhdocseinocore_moduleseino_adkagent_interface-check>
                                                                <label for=m-zhdocseinocore_moduleseino_adkagent_interface-check>
                                                                    <a href=/zh/docs/eino/core_modules/eino_adk/agent_interface/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkagent_interface>
                                                                        <span>Agent 抽象</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkagent_collaboration-li>
                                                                <input type=checkbox id=m-zhdocseinocore_moduleseino_adkagent_collaboration-check>
                                                                <label for=m-zhdocseinocore_moduleseino_adkagent_collaboration-check>
                                                                    <a href=/zh/docs/eino/core_modules/eino_adk/agent_collaboration/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkagent_collaboration>
                                                                        <span>Agent 协作</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section with-child" id=m-zhdocseinocore_moduleseino_adkagent_implementation-li>
                                                                <input type=checkbox id=m-zhdocseinocore_moduleseino_adkagent_implementation-check>
                                                                <label for=m-zhdocseinocore_moduleseino_adkagent_implementation-check>
                                                                    <a href=/zh/docs/eino/core_modules/eino_adk/agent_implementation/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinocore_moduleseino_adkagent_implementation>
                                                                        <span>Agent 实现</span>
                                                                    </a>
                                                                </label>
                                                                <ul class="ul-5 foldable">
                                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkagent_implementationchat_model-li>
                                                                        <input type=checkbox id=m-zhdocseinocore_moduleseino_adkagent_implementationchat_model-check>
                                                                        <label for=m-zhdocseinocore_moduleseino_adkagent_implementationchat_model-check>
                                                                            <a href=/zh/docs/eino/core_modules/eino_adk/agent_implementation/chat_model/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkagent_implementationchat_model>
                                                                                <span>ChatModelAgent</span>
                                                                            </a>
                                                                        </label>
                                                                    </li>
                                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkagent_implementationworkflow-li>
                                                                        <input type=checkbox id=m-zhdocseinocore_moduleseino_adkagent_implementationworkflow-check>
                                                                        <label for=m-zhdocseinocore_moduleseino_adkagent_implementationworkflow-check>
                                                                            <a href=/zh/docs/eino/core_modules/eino_adk/agent_implementation/workflow/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkagent_implementationworkflow>
                                                                                <span>Workflow Agents</span>
                                                                            </a>
                                                                        </label>
                                                                    </li>
                                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkagent_implementationsupervisor-li>
                                                                        <input type=checkbox id=m-zhdocseinocore_moduleseino_adkagent_implementationsupervisor-check>
                                                                        <label for=m-zhdocseinocore_moduleseino_adkagent_implementationsupervisor-check>
                                                                            <a href=/zh/docs/eino/core_modules/eino_adk/agent_implementation/supervisor/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkagent_implementationsupervisor>
                                                                                <span>Supervisor Agent</span>
                                                                            </a>
                                                                        </label>
                                                                    </li>
                                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkagent_implementationplan_execute-li>
                                                                        <input type=checkbox id=m-zhdocseinocore_moduleseino_adkagent_implementationplan_execute-check>
                                                                        <label for=m-zhdocseinocore_moduleseino_adkagent_implementationplan_execute-check>
                                                                            <a href=/zh/docs/eino/core_modules/eino_adk/agent_implementation/plan_execute/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkagent_implementationplan_execute>
                                                                                <span>Plan-Execute Agent</span>
                                                                            </a>
                                                                        </label>
                                                                    </li>
                                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkagent_implementationdeepagents-li>
                                                                        <input type=checkbox id=m-zhdocseinocore_moduleseino_adkagent_implementationdeepagents-check>
                                                                        <label for=m-zhdocseinocore_moduleseino_adkagent_implementationdeepagents-check>
                                                                            <a href=/zh/docs/eino/core_modules/eino_adk/agent_implementation/deepagents/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkagent_implementationdeepagents>
                                                                                <span>DeepAgents</span>
                                                                            </a>
                                                                        </label>
                                                                    </li>
                                                                </ul>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkagent_extension-li>
                                                                <input type=checkbox id=m-zhdocseinocore_moduleseino_adkagent_extension-check>
                                                                <label for=m-zhdocseinocore_moduleseino_adkagent_extension-check>
                                                                    <a href=/zh/docs/eino/core_modules/eino_adk/agent_extension/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkagent_extension>
                                                                        <span>Agent Runner 与扩展</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkagent_hitl-li>
                                                                <input type=checkbox id=m-zhdocseinocore_moduleseino_adkagent_hitl-check>
                                                                <label for=m-zhdocseinocore_moduleseino_adkagent_hitl-check>
                                                                    <a href=/zh/docs/eino/core_modules/eino_adk/agent_hitl/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkagent_hitl>
                                                                        <span>Eino human-in-the-loop框架：技术架构指南</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section with-child" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddleware-li>
                                                                <input type=checkbox id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddleware-check>
                                                                <label for=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddleware-check>
                                                                    <a href=/zh/docs/eino/core_modules/eino_adk/eino_adk_chatmodelagentmiddleware/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddleware>
                                                                        <span>ChatModelAgentMiddleware</span>
                                                                    </a>
                                                                </label>
                                                                <ul class="ul-5 foldable">
                                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section with-child" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewarefilesystem_backend-li>
                                                                        <input type=checkbox id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewarefilesystem_backend-check>
                                                                        <label for=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewarefilesystem_backend-check>
                                                                            <a href=/zh/docs/eino/core_modules/eino_adk/eino_adk_chatmodelagentmiddleware/filesystem_backend/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewarefilesystem_backend>
                                                                                <span>FileSystem Backend</span>
                                                                            </a>
                                                                        </label>
                                                                        <ul class="ul-6 foldable">
                                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewarefilesystem_backendbackend_ark_agentkit_sandbox-li>
                                                                                <input type=checkbox id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewarefilesystem_backendbackend_ark_agentkit_sandbox-check>
                                                                                <label for=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewarefilesystem_backendbackend_ark_agentkit_sandbox-check>
                                                                                    <a href=/zh/docs/eino/core_modules/eino_adk/eino_adk_chatmodelagentmiddleware/filesystem_backend/backend_ark_agentkit_sandbox/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewarefilesystem_backendbackend_ark_agentkit_sandbox>
                                                                                        <span>Ark Agentkit Sandbox</span>
                                                                                    </a>
                                                                                </label>
                                                                            </li>
                                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewarefilesystem_backendbackend_e69cace59cb0e69687e4bbb6e7b3bbe7bb9f-li>
                                                                                <input type=checkbox id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewarefilesystem_backendbackend_e69cace59cb0e69687e4bbb6e7b3bbe7bb9f-check>
                                                                                <label for=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewarefilesystem_backendbackend_e69cace59cb0e69687e4bbb6e7b3bbe7bb9f-check>
                                                                                    <a href=/zh/docs/eino/core_modules/eino_adk/eino_adk_chatmodelagentmiddleware/filesystem_backend/backend_%E6%9C%AC%E5%9C%B0%E6%96%87%E4%BB%B6%E7%B3%BB%E7%BB%9F/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewarefilesystem_backendbackend_e69cace59cb0e69687e4bbb6e7b3bbe7bb9f>
                                                                                        <span>本地文件系统</span>
                                                                                    </a>
                                                                                </label>
                                                                            </li>
                                                                        </ul>
                                                                    </li>
                                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_filesystem-li>
                                                                        <input type=checkbox id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_filesystem-check>
                                                                        <label for=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_filesystem-check>
                                                                            <a href=/zh/docs/eino/core_modules/eino_adk/eino_adk_chatmodelagentmiddleware/middleware_filesystem/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_filesystem>
                                                                                <span>FileSystem</span>
                                                                            </a>
                                                                        </label>
                                                                    </li>
                                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_skill-li>
                                                                        <input type=checkbox id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_skill-check>
                                                                        <label for=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_skill-check>
                                                                            <a href=/zh/docs/eino/core_modules/eino_adk/eino_adk_chatmodelagentmiddleware/middleware_skill/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_skill>
                                                                                <span>Skill</span>
                                                                            </a>
                                                                        </label>
                                                                    </li>
                                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_summarization-li>
                                                                        <input type=checkbox id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_summarization-check>
                                                                        <label for=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_summarization-check>
                                                                            <a href=/zh/docs/eino/core_modules/eino_adk/eino_adk_chatmodelagentmiddleware/middleware_summarization/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_summarization>
                                                                                <span>Summarization</span>
                                                                            </a>
                                                                        </label>
                                                                    </li>
                                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_toolreduction-li>
                                                                        <input type=checkbox id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_toolreduction-check>
                                                                        <label for=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_toolreduction-check>
                                                                            <a href=/zh/docs/eino/core_modules/eino_adk/eino_adk_chatmodelagentmiddleware/middleware_toolreduction/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_toolreduction>
                                                                                <span>Reduction</span>
                                                                            </a>
                                                                        </label>
                                                                    </li>
                                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_plantask-li>
                                                                        <input type=checkbox id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_plantask-check>
                                                                        <label for=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_plantask-check>
                                                                            <a href=/zh/docs/eino/core_modules/eino_adk/eino_adk_chatmodelagentmiddleware/middleware_plantask/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_plantask>
                                                                                <span>PlanTask</span>
                                                                            </a>
                                                                        </label>
                                                                    </li>
                                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_toolsearch-li>
                                                                        <input type=checkbox id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_toolsearch-check>
                                                                        <label for=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_toolsearch-check>
                                                                            <a href=/zh/docs/eino/core_modules/eino_adk/eino_adk_chatmodelagentmiddleware/middleware_toolsearch/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_toolsearch>
                                                                                <span>ToolSearch</span>
                                                                            </a>
                                                                        </label>
                                                                    </li>
                                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_patchtoolcalls-li>
                                                                        <input type=checkbox id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_patchtoolcalls-check>
                                                                        <label for=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_patchtoolcalls-check>
                                                                            <a href=/zh/docs/eino/core_modules/eino_adk/eino_adk_chatmodelagentmiddleware/middleware_patchtoolcalls/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_patchtoolcalls>
                                                                                <span>PatchToolCalls</span>
                                                                            </a>
                                                                        </label>
                                                                    </li>
                                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_agentsmd-li>
                                                                        <input type=checkbox id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_agentsmd-check>
                                                                        <label for=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_agentsmd-check>
                                                                            <a href=/zh/docs/eino/core_modules/eino_adk/eino_adk_chatmodelagentmiddleware/middleware_agentsmd/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkeino_adk_chatmodelagentmiddlewaremiddleware_agentsmd>
                                                                                <span>AgentsMD</span>
                                                                            </a>
                                                                        </label>
                                                                    </li>
                                                                </ul>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_moduleseino_adkadk_agent_callback-li>
                                                                <input type=checkbox id=m-zhdocseinocore_moduleseino_adkadk_agent_callback-check>
                                                                <label for=m-zhdocseinocore_moduleseino_adkadk_agent_callback-check>
                                                                    <a href=/zh/docs/eino/core_modules/eino_adk/adk_agent_callback/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_moduleseino_adkadk_agent_callback>
                                                                        <span>Agent Callback</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                        </ul>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section with-child" id=m-zhdocseinocore_modulesdevops-li>
                                                        <input type=checkbox id=m-zhdocseinocore_modulesdevops-check>
                                                        <label for=m-zhdocseinocore_modulesdevops-check>
                                                            <a href=/zh/docs/eino/core_modules/devops/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinocore_modulesdevops>
                                                                <span>应用开发工具链</span>
                                                            </a>
                                                        </label>
                                                        <ul class="ul-4 foldable">
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_modulesdevopside_plugin_guide-li>
                                                                <input type=checkbox id=m-zhdocseinocore_modulesdevopside_plugin_guide-check>
                                                                <label for=m-zhdocseinocore_modulesdevopside_plugin_guide-check>
                                                                    <a href=/zh/docs/eino/core_modules/devops/ide_plugin_guide/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_modulesdevopside_plugin_guide>
                                                                        <span>Eino Dev 插件安装指南</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_modulesdevopsvisual_orchestration_plugin_guide-li>
                                                                <input type=checkbox id=m-zhdocseinocore_modulesdevopsvisual_orchestration_plugin_guide-check>
                                                                <label for=m-zhdocseinocore_modulesdevopsvisual_orchestration_plugin_guide-check>
                                                                    <a href=/zh/docs/eino/core_modules/devops/visual_orchestration_plugin_guide/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_modulesdevopsvisual_orchestration_plugin_guide>
                                                                        <span>Eino Dev 可视化编排插件功能指南</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinocore_modulesdevopsvisual_debug_plugin_guide-li>
                                                                <input type=checkbox id=m-zhdocseinocore_modulesdevopsvisual_debug_plugin_guide-check>
                                                                <label for=m-zhdocseinocore_modulesdevopsvisual_debug_plugin_guide-check>
                                                                    <a href=/zh/docs/eino/core_modules/devops/visual_debug_plugin_guide/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinocore_modulesdevopsvisual_debug_plugin_guide>
                                                                        <span>Eino Dev 可视化调试插件功能指南</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                        </ul>
                                                    </li>
                                                </ul>
                                            </li>
                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section with-child" id=m-zhdocseinoecosystem_integration-li>
                                                <input type=checkbox id=m-zhdocseinoecosystem_integration-check checked>
                                                <label for=m-zhdocseinoecosystem_integration-check>
                                                    <a href=/zh/docs/eino/ecosystem_integration/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinoecosystem_integration>
                                                        <span>组件集成</span>
                                                    </a>
                                                </label>
                                                <ul class="ul-3 foldable">
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section with-child" id=m-zhdocseinoecosystem_integrationchat_model-li>
                                                        <input type=checkbox id=m-zhdocseinoecosystem_integrationchat_model-check>
                                                        <label for=m-zhdocseinoecosystem_integrationchat_model-check>
                                                            <a href=/zh/docs/eino/ecosystem_integration/chat_model/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinoecosystem_integrationchat_model>
                                                                <span>ChatModel</span>
                                                            </a>
                                                        </label>
                                                        <ul class="ul-4 foldable">
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoecosystem_integrationchat_modelagentic_model_openai-li>
                                                                <input type=checkbox id=m-zhdocseinoecosystem_integrationchat_modelagentic_model_openai-check>
                                                                <label for=m-zhdocseinoecosystem_integrationchat_modelagentic_model_openai-check>
                                                                    <a href=/zh/docs/eino/ecosystem_integration/chat_model/agentic_model_openai/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinoecosystem_integrationchat_modelagentic_model_openai>
                                                                        <span>OpenAI</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoecosystem_integrationchat_modelagentic_model_ark-li>
                                                                <input type=checkbox id=m-zhdocseinoecosystem_integrationchat_modelagentic_model_ark-check>
                                                                <label for=m-zhdocseinoecosystem_integrationchat_modelagentic_model_ark-check>
                                                                    <a href=/zh/docs/eino/ecosystem_integration/chat_model/agentic_model_ark/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinoecosystem_integrationchat_modelagentic_model_ark>
                                                                        <span>ARK</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                        </ul>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoecosystem_integrationdocument-li>
                                                        <input type=checkbox id=m-zhdocseinoecosystem_integrationdocument-check>
                                                        <label for=m-zhdocseinoecosystem_integrationdocument-check>
                                                            <a href=/zh/docs/eino/ecosystem_integration/document/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinoecosystem_integrationdocument>
                                                                <span>Document</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoecosystem_integrationembedding-li>
                                                        <input type=checkbox id=m-zhdocseinoecosystem_integrationembedding-check>
                                                        <label for=m-zhdocseinoecosystem_integrationembedding-check>
                                                            <a href=/zh/docs/eino/ecosystem_integration/embedding/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinoecosystem_integrationembedding>
                                                                <span>Embedding</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoecosystem_integrationtool-li>
                                                        <input type=checkbox id=m-zhdocseinoecosystem_integrationtool-check>
                                                        <label for=m-zhdocseinoecosystem_integrationtool-check>
                                                            <a href=/zh/docs/eino/ecosystem_integration/tool/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinoecosystem_integrationtool>
                                                                <span>Tool</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoecosystem_integrationcallbacks-li>
                                                        <input type=checkbox id=m-zhdocseinoecosystem_integrationcallbacks-check>
                                                        <label for=m-zhdocseinoecosystem_integrationcallbacks-check>
                                                            <a href=/zh/docs/eino/ecosystem_integration/callbacks/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinoecosystem_integrationcallbacks>
                                                                <span>Callbacks</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoecosystem_integrationindexer-li>
                                                        <input type=checkbox id=m-zhdocseinoecosystem_integrationindexer-check>
                                                        <label for=m-zhdocseinoecosystem_integrationindexer-check>
                                                            <a href=/zh/docs/eino/ecosystem_integration/indexer/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinoecosystem_integrationindexer>
                                                                <span>Indexer</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoecosystem_integrationretriever-li>
                                                        <input type=checkbox id=m-zhdocseinoecosystem_integrationretriever-check>
                                                        <label for=m-zhdocseinoecosystem_integrationretriever-check>
                                                            <a href=/zh/docs/eino/ecosystem_integration/retriever/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinoecosystem_integrationretriever>
                                                                <span>Retriever</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinoecosystem_integrationchat_template-li>
                                                        <input type=checkbox id=m-zhdocseinoecosystem_integrationchat_template-check>
                                                        <label for=m-zhdocseinoecosystem_integrationchat_template-check>
                                                            <a href=/zh/docs/eino/ecosystem_integration/chat_template/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinoecosystem_integrationchat_template>
                                                                <span>ChatTemplate</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                </ul>
                                            </li>
                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section with-child" id=m-zhdocseinorelease_notes_and_migration-li>
                                                <input type=checkbox id=m-zhdocseinorelease_notes_and_migration-check checked>
                                                <label for=m-zhdocseinorelease_notes_and_migration-check>
                                                    <a href=/zh/docs/eino/release_notes_and_migration/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinorelease_notes_and_migration>
                                                        <span>发布记录 & 迁移指引</span>
                                                    </a>
                                                </label>
                                                <ul class="ul-3 foldable">
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinorelease_notes_and_migrationv01_first_release-li>
                                                        <input type=checkbox id=m-zhdocseinorelease_notes_and_migrationv01_first_release-check>
                                                        <label for=m-zhdocseinorelease_notes_and_migrationv01_first_release-check>
                                                            <a href=/zh/docs/eino/release_notes_and_migration/v01_first_release/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinorelease_notes_and_migrationv01_first_release>
                                                                <span>v0.1.*-first release</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinorelease_notes_and_migrationv02_second_release-li>
                                                        <input type=checkbox id=m-zhdocseinorelease_notes_and_migrationv02_second_release-check>
                                                        <label for=m-zhdocseinorelease_notes_and_migrationv02_second_release-check>
                                                            <a href=/zh/docs/eino/release_notes_and_migration/v02_second_release/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinorelease_notes_and_migrationv02_second_release>
                                                                <span>v0.2.*-second release</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinorelease_notes_and_migrationv03_tiny_break_change-li>
                                                        <input type=checkbox id=m-zhdocseinorelease_notes_and_migrationv03_tiny_break_change-check>
                                                        <label for=m-zhdocseinorelease_notes_and_migrationv03_tiny_break_change-check>
                                                            <a href=/zh/docs/eino/release_notes_and_migration/v03_tiny_break_change/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinorelease_notes_and_migrationv03_tiny_break_change>
                                                                <span>v0.3.*-tiny break change</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinorelease_notes_and_migrationeino_v04_-compose_optimization-li>
                                                        <input type=checkbox id=m-zhdocseinorelease_notes_and_migrationeino_v04_-compose_optimization-check>
                                                        <label for=m-zhdocseinorelease_notes_and_migrationeino_v04_-compose_optimization-check>
                                                            <a href=/zh/docs/eino/release_notes_and_migration/eino_v0.4._-compose_optimization/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinorelease_notes_and_migrationeino_v04_-compose_optimization>
                                                                <span>v0.4.*-compose optimization</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinorelease_notes_and_migrationeino_v05_-adk_implementation-li>
                                                        <input type=checkbox id=m-zhdocseinorelease_notes_and_migrationeino_v05_-adk_implementation-check>
                                                        <label for=m-zhdocseinorelease_notes_and_migrationeino_v05_-adk_implementation-check>
                                                            <a href=/zh/docs/eino/release_notes_and_migration/eino_v0.5._-adk_implementation/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinorelease_notes_and_migrationeino_v05_-adk_implementation>
                                                                <span>v0.5.*-ADK implementation</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinorelease_notes_and_migrationeino_v06_-jsonschema_optimization-li>
                                                        <input type=checkbox id=m-zhdocseinorelease_notes_and_migrationeino_v06_-jsonschema_optimization-check>
                                                        <label for=m-zhdocseinorelease_notes_and_migrationeino_v06_-jsonschema_optimization-check>
                                                            <a href=/zh/docs/eino/release_notes_and_migration/eino_v0.6._-jsonschema_optimization/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinorelease_notes_and_migrationeino_v06_-jsonschema_optimization>
                                                                <span>v0.6.*-jsonschema optimization</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinorelease_notes_and_migrationeino_v07_-interrupt_resume_refactor-li>
                                                        <input type=checkbox id=m-zhdocseinorelease_notes_and_migrationeino_v07_-interrupt_resume_refactor-check>
                                                        <label for=m-zhdocseinorelease_notes_and_migrationeino_v07_-interrupt_resume_refactor-check>
                                                            <a href=/zh/docs/eino/release_notes_and_migration/eino_v0.7._-interrupt_resume_refactor/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinorelease_notes_and_migrationeino_v07_-interrupt_resume_refactor>
                                                                <span>v0.7.*-interrupt resume refactor</span>
                                                            </a>
                                                        </label>
                                                    </li>
                                                    <li class="td-sidebar-nav__section-title td-sidebar-nav__section with-child" id=m-zhdocseinorelease_notes_and_migrationeino_v08_-adk_middlewares-li>
                                                        <input type=checkbox id=m-zhdocseinorelease_notes_and_migrationeino_v08_-adk_middlewares-check>
                                                        <label for=m-zhdocseinorelease_notes_and_migrationeino_v08_-adk_middlewares-check>
                                                            <a href=/zh/docs/eino/release_notes_and_migration/eino_v0.8._-adk_middlewares/ class="align-left ps-0 td-sidebar-link td-sidebar-link__section" id=m-zhdocseinorelease_notes_and_migrationeino_v08_-adk_middlewares>
                                                                <span>v0.8.*-adk middlewares</span>
                                                            </a>
                                                        </label>
                                                        <ul class="ul-4 foldable">
                                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinorelease_notes_and_migrationeino_v08_-adk_middlewareseino_v08_e4b88de585bce5aeb9e69bb4e696b0-li>
                                                                <input type=checkbox id=m-zhdocseinorelease_notes_and_migrationeino_v08_-adk_middlewareseino_v08_e4b88de585bce5aeb9e69bb4e696b0-check>
                                                                <label for=m-zhdocseinorelease_notes_and_migrationeino_v08_-adk_middlewareseino_v08_e4b88de585bce5aeb9e69bb4e696b0-check>
                                                                    <a href=/zh/docs/eino/release_notes_and_migration/eino_v0.8._-adk_middlewares/eino_v0.8_%E4%B8%8D%E5%85%BC%E5%AE%B9%E6%9B%B4%E6%96%B0/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinorelease_notes_and_migrationeino_v08_-adk_middlewareseino_v08_e4b88de585bce5aeb9e69bb4e696b0>
                                                                        <span>Eino v0.8 不兼容更新</span>
                                                                    </a>
                                                                </label>
                                                            </li>
                                                        </ul>
                                                    </li>
                                                </ul>
                                            </li>
                                            <li class="td-sidebar-nav__section-title td-sidebar-nav__section without-child" id=m-zhdocseinofaq-li>
                                                <input type=checkbox id=m-zhdocseinofaq-check checked>
                                                <label for=m-zhdocseinofaq-check>
                                                    <a href=/zh/docs/eino/faq/ class="align-left ps-0 td-sidebar-link td-sidebar-link__page" id=m-zhdocseinofaq>
                                                        <span>FAQ</span>
                                                    </a>
                                                </label>
                                            </li>
                                        </ul>
                                    </li>
                                </ul>
                            </li>
                        </ul>
                    </nav>
                </div>
            </aside>
            <aside class="d-none d-xl-block col-xl-2 td-sidebar-toc d-print-none">
                <div class="td-page-meta ms-2 pb-1 pt-2 mb-0">
                    <a href=https://github.com/cloudwego/cloudwego.github.io/edit/main/content/zh/docs/eino/quick_start/chapter_05_middleware.md class="td-page-meta--edit td-page-meta__edit" target=_blank rel=noopener>
                        <i class="fa-solid fa-pen-to-square fa-fw"></i>
                        编辑此页
                    </a>
                    <a href="https://github.com/cloudwego/cloudwego.github.io/new/main/content/zh/docs/eino/quick_start?filename=change-me.md&amp;value=---%0Atitle%3A+%22Long+Page+Title%22%0AlinkTitle%3A+%22Short+Nav+Title%22%0Aweight%3A+100%0Adescription%3A+%3E-%0A+++++Page+description+for+heading+and+indexes.%0A---%0A%0A%23%23+Heading%0A%0AEdit+this+template+to+create+your+new+page.%0A%0A%2A+Give+it+a+good+name%2C+ending+in+%60.md%60+-+e.g.+%60getting-started.md%60%0A%2A+Edit+the+%22front+matter%22+section+at+the+top+of+the+page+%28weight+controls+how+its+ordered+amongst+other+pages+in+the+same+directory%3B+lowest+number+first%29.%0A%2A+Add+a+good+commit+message+at+the+bottom+of+the+page+%28%3C80+characters%3B+use+the+extended+description+field+for+more+detail%29.%0A%2A+Create+a+new+branch+so+you+can+preview+your+new+file+and+request+a+review+via+Pull+Request.%0A" class="td-page-meta--child td-page-meta__child" target=_blank rel=noopener>
                        <i class="fa-solid fa-pen-to-square fa-fw"></i>
                        添加子页面
                    </a>
                    <a href="https://github.com/cloudwego/cloudwego.github.io/issues/new?title=%e7%ac%ac%e4%ba%94%e7%ab%a0%ef%bc%9aMiddleware%ef%bc%88%e4%b8%ad%e9%97%b4%e4%bb%b6%e6%a8%a1%e5%bc%8f%ef%bc%89" class="td-page-meta--issue td-page-meta__issue" target=_blank rel=noopener>
                        <i class="fa-brands fa-github fa-fw"></i>
                        提交文档问题
                    </a>
                    <a href=https://github.com/cloudwego/eino/issues/new/choose class="td-page-meta--project td-page-meta__project-issue" target=_blank rel=noopener>
                        <i class="fa-solid fa-list-check fa-fw"></i>
                        提交项目问题
                    </a>
                </div>
                <div class=td-toc>
                    <nav id=TableOfContents>
                        <ul>
                            <li>
                                <a href=#为什么需要-middleware>为什么需要 Middleware</a>
                                <ul>
                                    <li>
                                        <a href=#问题一tool-错误会中断整个流程>问题一：Tool 错误会中断整个流程</a>
                                    </li>
                                    <li>
                                        <a href=#问题二模型调用可能因限流失败>问题二：模型调用可能因限流失败</a>
                                    </li>
                                    <li>
                                        <a href=#期望的行为>期望的行为</a>
                                    </li>
                                    <li>
                                        <a href=#middleware-的定位>Middleware 的定位</a>
                                    </li>
                                </ul>
                            </li>
                            <li>
                                <a href=#代码位置>代码位置</a>
                            </li>
                            <li>
                                <a href=#前置条件>前置条件</a>
                            </li>
                            <li>
                                <a href=#运行>运行</a>
                            </li>
                            <li>
                                <a href=#关键概念>关键概念</a>
                                <ul>
                                    <li>
                                        <a href=#middleware-接口>Middleware 接口</a>
                                    </li>
                                    <li>
                                        <a href=#middleware-执行顺序>Middleware 执行顺序</a>
                                    </li>
                                    <li>
                                        <a href=#safetoolmiddleware>SafeToolMiddleware</a>
                                    </li>
                                    <li>
                                        <a href=#modelretryconfig>ModelRetryConfig</a>
                                    </li>
                                </ul>
                            </li>
                            <li>
                                <a href=#middleware-的实现>Middleware 的实现</a>
                                <ul>
                                    <li>
                                        <a href=#1-实现-safetoolmiddleware>1. 实现 SafeToolMiddleware</a>
                                    </li>
                                    <li>
                                        <a href=#2-实现流式-tool-错误处理>2. 实现流式 Tool 错误处理</a>
                                    </li>
                                    <li>
                                        <a href=#3-配置-agent-使用-middleware>3. 配置 Agent 使用 Middleware</a>
                                    </li>
                                </ul>
                            </li>
                            <li>
                                <a href=#middleware-执行流程>Middleware 执行流程</a>
                            </li>
                            <li>
                                <a href=#本章小结>本章小结</a>
                            </li>
                            <li>
                                <a href=#扩展思考>扩展思考</a>
                            </li>
                        </ul>
                    </nav>
                </div>
            </aside>
            <main class="col-12 col-md-9 col-xl-8 ps-md-5" role=main>
                <nav aria-label=breadcrumb class="d-none d-md-block d-print-none">
                    <ol class="breadcrumb spb-1">
                        <li class=breadcrumb-item>
                            <a href=https://www.cloudwego.io/zh/docs/>文档</a>
                        </li>
                        <li class=breadcrumb-item>
                            <a href=https://www.cloudwego.io/zh/docs/eino/>Eino</a>
                        </li>
                        <li class=breadcrumb-item>
                            <a href=https://www.cloudwego.io/zh/docs/eino/quick_start/>快速开始</a>
                        </li>
                        <li class="breadcrumb-item active" aria-current=page>
                            <a href=https://www.cloudwego.io/zh/docs/eino/quick_start/chapter_05_middleware/>第五章：Middleware（中间件模式）</a>
                        </li>
                    </ol>
                </nav>
                <div class=td-content>
                    <div class="copy-fulltext btn-group" id=copy-fulltext data-title=第五章：Middleware（中间件模式） data-url=https://www.cloudwego.io/zh/docs/eino/quick_start/chapter_05_middleware/ data-success-markdown="已复制 Markdown 格式" data-success-text=已复制纯文本>
                        <button type=button class="btn btn-sm btn-outline-secondary" id=copy-fulltext-default title=复制全文>
                            <i class="fa-regular fa-copy"></i>
                            <span>复制全文</span>
                        </button>
                        <button type=button class="btn btn-sm btn-outline-secondary dropdown-toggle dropdown-toggle-split" data-bs-toggle=dropdown aria-haspopup=true aria-expanded=false>
                            <span class=visually-hidden>Toggle Dropdown</span>
                        </button>
                        <div class="dropdown-menu dropdown-menu-end">
                            <button class=dropdown-item data-copy-type=markdown>复制 Markdown
</button>
                            <button class=dropdown-item data-copy-type=text>复制文本</button>
                        </div>
                        <div class=copy-fulltext__toast id=copy-fulltext-toast></div>
                    </div>
                    <script id=copy-fulltext-markdown type=text/plain>

本章目标：理解 Middleware 模式，实现 Tool 错误处理和 ChatModel 重试机制。

## 为什么需要 Middleware

第四章我们为 Agent 添加了 Tool 能力，让 Agent 能够访问文件系统。但在实际应用场景中，**Tool 报错或 ChatModel 报错是常见的现象**，例如：

- **Tool 报错**：文件不存在、参数错误、权限不足等
- **ChatModel 报错**：API 限流（429）、网络超时、服务不可用等

### 问题一：Tool 错误会中断整个流程

当 Tool 执行失败时，错误会直接传播到 Agent，导致整个对话中断：


[tool call] read_file(file_path: "nonexistent.txt")
Error: open nonexistent.txt: no such file or directory
// 对话中断，用户需要重新开始


### 问题二：模型调用可能因限流失败

当模型 API 返回 429（Too Many Requests）错误时，整个对话也会中断：

Error: rate limit exceeded (429)
// 对话中断

### 期望的行为

这些报错信息往往**不希望直接终止 Agent 流程**，而是希望把报错信息给到模型，由模型自动纠错进行下一轮。例如：

[tool call] read_file(file_path: "nonexistent.txt")
[tool result] [tool error] open nonexistent.txt: no such file or directory
[assistant] 抱歉，文件不存在。让我先列出当前目录的文件...
[tool call] glob(pattern: "*")

### Middleware 的定位


## Middleware 的实现

### 1. 实现 SafeToolMiddleware

### 2. 实现流式 Tool 错误处理

### 3. 配置 Agent 使用 Middleware


func (m *safeToolMiddleware) WrapInvokableToolCall(
    _ context.Context,
    endpoint adk.InvokableToolCallEndpoint,
    _ *adk.ToolContext,
) (adk.InvokableToolCallEndpoint, error) {
    return func(ctx context.Context, args string, opts ...tool.Option) (string, error) {
        result, err := endpoint(ctx, args, opts...)
        if err != nil {
            if _, ok := compose.IsInterruptRerunError(err); ok {
                return "", err
            }
            return fmt.Sprintf("[tool error] %v", err), nil
        }
        return result, nil
    }, nil
}

// 配置 DeepAgent（与第四章一样，新增 Handlers 和 ModelRetryConfig）
agent, _ := deep.New(ctx, &deep.Config{
    ChatModel:      cm,
    Backend:        backend,
    StreamingShell: backend,
    MaxIteration:   50,
    Handlers: []adk.ChatModelAgentMiddleware{
        &safeToolMiddleware{},
    },
    ModelRetryConfig: &adk.ModelRetryConfig{
        MaxRetries: 5,
        IsRetryAble: func(_ context.Context, err error) bool {
            return strings.Contains(err.Error(), "429")
        },
    },
})
┌─────────────────────────────────────────┐
│  用户：读取不存在的文件                   │
└─────────────────────────────────────────┘
                   ↓
        ┌──────────────────────┐
        │  Agent 分析意图       │
        │  决定调用 read_file   │
        └──────────────────────┘
                   ↓
        ┌──────────────────────┐
        │  SafeToolMiddleware  │
        │  拦截 Tool 调用       │
        └──────────────────────┘
                   ↓
        ┌──────────────────────┐
        │  执行 read_file       │
        │  返回错误             │
        └──────────────────────┘
                   ↓
        ┌──────────────────────┐
        │  SafeToolMiddleware  │
        │  将错误转换为字符串    │
        └──────────────────────┘
                   ↓
        ┌──────────────────────┐
        │  返回 Tool Result     │
        │  "[tool error] ..."   │
        └──────────────────────┘
                   ↓
        ┌──────────────────────┐
        │  Agent 生成回复       │
        │  "抱歉，文件不存在..." │
        └──────────────────────┘

## 本章小结

- **Middleware**：Agent 的拦截器，可以在调用前后插入自定义逻辑
- **SafeToolMiddleware**：将 Tool 错误转换为字符串，让模型能够理解并处理
- **ModelRetryConfig**：配置 ChatModel 的自动重试，处理限流等临时错误
- **装饰器模式**：Middleware 包装原始调用，可以修改输入、输出或错误
- **洋葱模型**：请求从外向内穿过 Middleware，响应从内向外返回

## 扩展思考

**Eino 内置 Middleware：**

<table>
<tr><td>Middleware</td><td>功能说明</td></tr>
<tr><td><strong>reduction</strong></td><td>工具输出缩减，当工具返回内容过长时自动截断并卸载到文件系统，防止上下文溢出</td></tr>
<tr><td><strong>summarization</strong></td><td>对话历史自动摘要，当 token 数量超过阈值时自动生成摘要压缩历史</td></tr>
<tr><td><strong>skill</strong></td><td>技能加载中间件，让 Agent 能够动态加载和执行预定义的技能</td></tr>
</table>

**Middleware 链示例：**

                    </script>
                    <h1>第五章：Middleware（中间件模式）</h1>
                    <header class=article-meta></header>
                    <p>本章目标：理解 Middleware 模式，实现 Tool 错误处理和 ChatModel 重试机制。</p>
                    <h2 id=为什么需要-middleware>为什么需要 Middleware</h2>
                    <p>
                        第四章我们为 Agent 添加了 Tool 能力，让 Agent 能够访问文件系统。但在实际应用场景中，<strong>Tool 报错或 ChatModel 报错是常见的现象</strong>
                        ，例如：
                    </p>
                    <ul>
                        <li>
                            <strong>Tool 报错</strong>
                            ：文件不存在、参数错误、权限不足等
                        </li>
                        <li>
                            <strong>ChatModel 报错</strong>
                            ：API 限流（429）、网络超时、服务不可用等
                        </li>
                    </ul>
                    <h3 id=问题一tool-错误会中断整个流程>问题一：Tool 错误会中断整个流程</h3>
                    <p>当 Tool 执行失败时，错误会直接传播到 Agent，导致整个对话中断：</p>
                    <pre tabindex=0>
                        <code>[tool call] read_file(file_path: &#34;nonexistent.txt &#34;)
Error: open nonexistent.txt: no such file or directory
// 对话中断，用户需要重新开始
</code>
                    </pre>
                    <h3 id=问题二模型调用可能因限流失败>问题二：模型调用可能因限流失败</h3>
                    <p>当模型 API 返回 429（Too Many Requests）错误时，整个对话也会中断：</p>
                    <pre tabindex=0>
                        <code>Error: rate limit exceeded (429)
// 对话中断
</code>
                    </pre>
                    <h3 id=期望的行为>期望的行为</h3>
                    <p>
                        这些报错信息往往<strong>不希望直接终止 Agent 流程</strong>
                        ，而是希望把报错信息给到模型，由模型自动纠错进行下一轮。例如：
                    </p>
                    <pre tabindex=0>
                        <code>[tool call] read_file(file_path: &#34;nonexistent.txt &#34;)
[tool result] [tool error] open nonexistent.txt: no such file or directory
[assistant] 抱歉，文件不存在。让我先列出当前目录的文件...
[tool call] glob(pattern: &#34;*&#34;)
</code>
                    </pre>
                    <h3 id=middleware-的定位>Middleware 的定位</h3>
                    <p>
                        <strong>Middleware 模式</strong>
                        可以扩展 Tool 和 ChatModel 的行为，非常适合解决这个问题：
                    </p>
                    <ul>
                        <li>
                            <strong>Middleware 是 Agent 的拦截器</strong>
                            ：在调用前后插入自定义逻辑
                        </li>
                        <li>
                            <strong>Middleware 可处理错误</strong>
                            ：将错误转换为模型可理解的格式
                        </li>
                        <li>
                            <strong>Middleware 可实现重试</strong>
                            ：自动重试失败的操作
                        </li>
                        <li>
                            <strong>Middleware 可组合</strong>
                            ：多个 Middleware 可以串联使用
                        </li>
                    </ul>
                    <p>
                        <strong>简单类比：</strong>
                    </p>
                    <ul>
                        <li>
                            <strong>Agent</strong>
                            = &ldquo;业务逻辑 &rdquo;
                        </li>
                        <li>
                            <strong>Middleware</strong>
                            = &ldquo;AOP 切面 &rdquo;（日志、重试、错误处理等横切关注点）
                        </li>
                    </ul>
                    <h2 id=代码位置>代码位置</h2>
                    <ul>
                        <li>
                            入口代码：<a href=https://github.com/cloudwego/eino-examples/blob/main/quickstart/chatwitheino/cmd/ch05/main.go>cmd/ch05/main.go</a>
                        </li>
                    </ul>
                    <h2 id=前置条件>前置条件</h2>
                    <p>
                        与第一章一致：需要配置一个可用的 ChatModel（OpenAI 或 Ark）。同时，需要与第四章一样设置 <code>PROJECT_ROOT</code>
                        ：
                    </p>
                    <div class=highlight>
                        <pre tabindex=0 class=chroma>
                            <code class=language-bash data-lang=bash>
                                <span class=line>
                                    <span class=cl>
                                        <span class=nb>export</span>
                                        <span class=nv>PROJECT_ROOT</span>
                                        <span class=o>=</span>
                                        /path/to/eino  <span class=c1># Eino 核心库根目录</span>
                                    </span>
                                </span>
                            </code>
                        </pre>
                    </div>
                    <h2 id=运行>运行</h2>
                    <p>
                        在 <code>examples/quickstart/chatwitheino</code>
                        目录下执行：
                    </p>
                    <div class=highlight>
                        <pre tabindex=0 class=chroma>
                            <code class=language-bash data-lang=bash>
                                <span class=line>
                                    <span class=cl>
                                        <span class=c1># 设置项目根目录</span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=nb>export</span>
                                        <span class=nv>PROJECT_ROOT</span>
                                        <span class=o>=</span>
                                        /path/to/your/project

                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl></span>
                                </span>
                                <span class=line>
                                    <span class=cl>go run ./cmd/ch05
</span>
                                </span>
                            </code>
                        </pre>
                    </div>
                    <p>输出示例：</p>
                    <pre tabindex=0>
                        <code>you &gt;列出当前目录的文件
[assistant] 我来帮你列出文件...
[tool call] list_files(directory: &#34;.&#34;)

you &gt;读取一个不存在的文件
[assistant] 尝试读取文件...
[tool call] read_file(file_path: &#34;nonexistent.txt &#34;)
[tool result] [tool error] open nonexistent.txt: no such file or directory
[assistant] 抱歉，文件不存在...
</code>
                    </pre>
                    <h2 id=关键概念>关键概念</h2>
                    <h3 id=middleware-接口>Middleware 接口</h3>
                    <p>
                        <code>ChatModelAgentMiddleware</code>
                        是 Agent 的中间件接口：
                    </p>
                    <div class=highlight>
                        <pre tabindex=0 class=chroma>
                            <code class=language-go data-lang=go>
                                <span class=line>
                                    <span class=cl>
                                        <span class=kd>type</span>
                                        <span class=w></span>
                                        <span class=nx>ChatModelAgentMiddleware</span>
                                        <span class=w></span>
                                        <span class=kd>interface</span>
                                        <span class=w></span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// BeforeAgent is called before each agent run, allowing modification of</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// the agent &#39;s instruction and tools configuration.</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nf>BeforeAgent</span>
                                        <span class=p>(</span>
                                        <span class=nx>ctx</span>
                                        <span class=w></span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>runCtx</span>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>ChatModelAgentContext</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>(</span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>ChatModelAgentContext</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=kt>error</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// BeforeModelRewriteState is called before each model invocation.</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// The returned state is persisted to the agent &#39;s internal state and passed to the model.</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nf>BeforeModelRewriteState</span>
                                        <span class=p>(</span>
                                        <span class=nx>ctx</span>
                                        <span class=w></span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>state</span>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>ChatModelAgentState</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>mc</span>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>ModelContext</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>(</span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>ChatModelAgentState</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=kt>error</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// AfterModelRewriteState is called after each model invocation.</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// The input state includes the model &#39;s response as the last message.</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nf>AfterModelRewriteState</span>
                                        <span class=p>(</span>
                                        <span class=nx>ctx</span>
                                        <span class=w></span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>state</span>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>ChatModelAgentState</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>mc</span>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>ModelContext</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>(</span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>ChatModelAgentState</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=kt>error</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// WrapInvokableToolCall wraps a tool &#39;s synchronous execution with custom behavior.</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// This method is only called for tools that implement InvokableTool.</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nf>WrapInvokableToolCall</span>
                                        <span class=p>(</span>
                                        <span class=nx>ctx</span>
                                        <span class=w></span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>endpoint</span>
                                        <span class=w></span>
                                        <span class=nx>InvokableToolCallEndpoint</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>tCtx</span>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>ToolContext</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>(</span>
                                        <span class=nx>InvokableToolCallEndpoint</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=kt>error</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// WrapStreamableToolCall wraps a tool &#39;s streaming execution with custom behavior.</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// This method is only called for tools that implement StreamableTool.</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nf>WrapStreamableToolCall</span>
                                        <span class=p>(</span>
                                        <span class=nx>ctx</span>
                                        <span class=w></span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>endpoint</span>
                                        <span class=w></span>
                                        <span class=nx>StreamableToolCallEndpoint</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>tCtx</span>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>ToolContext</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>(</span>
                                        <span class=nx>StreamableToolCallEndpoint</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=kt>error</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// WrapEnhancedInvokableToolCall wraps an enhanced tool &#39;s synchronous execution.</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// This method is only called for tools that implement EnhancedInvokableTool.</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nf>WrapEnhancedInvokableToolCall</span>
                                        <span class=p>(</span>
                                        <span class=nx>ctx</span>
                                        <span class=w></span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>endpoint</span>
                                        <span class=w></span>
                                        <span class=nx>EnhancedInvokableToolCallEndpoint</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>tCtx</span>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>ToolContext</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>(</span>
                                        <span class=nx>EnhancedInvokableToolCallEndpoint</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=kt>error</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// WrapEnhancedStreamableToolCall wraps an enhanced tool &#39;s streaming execution.</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// This method is only called for tools that implement EnhancedStreamableTool.</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nf>WrapEnhancedStreamableToolCall</span>
                                        <span class=p>(</span>
                                        <span class=nx>ctx</span>
                                        <span class=w></span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>endpoint</span>
                                        <span class=w></span>
                                        <span class=nx>EnhancedStreamableToolCallEndpoint</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>tCtx</span>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>ToolContext</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>(</span>
                                        <span class=nx>EnhancedStreamableToolCallEndpoint</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=kt>error</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// WrapModel wraps a chat model with custom behavior.</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// This method is called at request time when the model is about to be invoked.</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nf>WrapModel</span>
                                        <span class=p>(</span>
                                        <span class=nx>ctx</span>
                                        <span class=w></span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>m</span>
                                        <span class=w></span>
                                        <span class=nx>model</span>
                                        <span class=p>.</span>
                                        <span class=nx>BaseChatModel</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>mc</span>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>ModelContext</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>(</span>
                                        <span class=nx>model</span>
                                        <span class=p>.</span>
                                        <span class=nx>BaseChatModel</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=kt>error</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>}</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                            </code>
                        </pre>
                    </div>
                    <p>
                        <strong>设计理念：</strong>
                    </p>
                    <ul>
                        <li>
                            <strong>装饰器模式</strong>
                            ：每个 Middleware 包装原始调用，可以修改输入、输出或错误
                        </li>
                        <li>
                            <strong>洋葱模型</strong>
                            ：请求从外向内穿过 Middleware，响应从内向外返回
                        </li>
                        <li>
                            <strong>可组合</strong>
                            ：多个 Middleware 按顺序执行
                        </li>
                    </ul>
                    <h3 id=middleware-执行顺序>Middleware 执行顺序</h3>
                    <p>
                        <code>Handlers</code>
                        （即 Middlewares）按<strong>数组正序</strong>
                        包装，形成洋葱模型：
                    </p>
                    <div class=highlight>
                        <pre tabindex=0 class=chroma>
                            <code class=language-go data-lang=go>
                                <span class=line>
                                    <span class=cl>
                                        <span class=nx>Handlers</span>
                                        <span class=p>:</span>
                                        <span class=w></span>
                                        <span class=p>[]</span>
                                        <span class=nx>adk</span>
                                        <span class=p>.</span>
                                        <span class=nx>ChatModelAgentMiddleware</span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=o>&amp;</span>
                                        <span class=nx>middlewareA</span>
                                        <span class=p>{},</span>
                                        <span class=w></span>
                                        <span class=c1>// 最外层：最先 Wrap，最先拦截请求，但 WrapModel 最后生效</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=o>&amp;</span>
                                        <span class=nx>middlewareB</span>
                                        <span class=p>{},</span>
                                        <span class=w></span>
                                        <span class=c1>// 中间层</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=o>&amp;</span>
                                        <span class=nx>middlewareC</span>
                                        <span class=p>{},</span>
                                        <span class=w></span>
                                        <span class=c1>// 最内层：最后 Wrap</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>}</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                            </code>
                        </pre>
                    </div>
                    <p>
                        <strong>对于 Tool 调用的执行顺序：</strong>
                    </p>
                    <pre tabindex=0>
                        <code>请求 → A.Wrap → B.Wrap → C.Wrap → 实际 Tool 执行 → C返回 → B返回 → A返回 → 响应
</code>
                    </pre>
                    <p>
                        <strong>实用建议：</strong>
                        将 <code>safeToolMiddleware</code>
                        （错误捕获）放在最内层（数组末尾），确保其他 Middleware 抛出的中断错误能正确向外传播。
                    </p>
                    <h3 id=safetoolmiddleware>SafeToolMiddleware</h3>
                    <p>
                        <code>SafeToolMiddleware</code>
                        将 Tool 错误转换为字符串，让模型能够理解并处理：
                    </p>
                    <div class=highlight>
                        <pre tabindex=0 class=chroma>
                            <code class=language-go data-lang=go>
                                <span class=line>
                                    <span class=cl>
                                        <span class=kd>type</span>
                                        <span class=w></span>
                                        <span class=nx>safeToolMiddleware</span>
                                        <span class=w></span>
                                        <span class=kd>struct</span>
                                        <span class=w></span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>adk</span>
                                        <span class=p>.</span>
                                        <span class=nx>BaseChatModelAgentMiddleware</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>}</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=kd>func</span>
                                        <span class=w></span>
                                        <span class=p>(</span>
                                        <span class=nx>m</span>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>safeToolMiddleware</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=nf>WrapInvokableToolCall</span>
                                        <span class=p>(</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>_</span>
                                        <span class=w></span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>endpoint</span>
                                        <span class=w></span>
                                        <span class=nx>adk</span>
                                        <span class=p>.</span>
                                        <span class=nx>InvokableToolCallEndpoint</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>_</span>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>adk</span>
                                        <span class=p>.</span>
                                        <span class=nx>ToolContext</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>(</span>
                                        <span class=nx>adk</span>
                                        <span class=p>.</span>
                                        <span class=nx>InvokableToolCallEndpoint</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=kt>error</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=k>return</span>
                                        <span class=w></span>
                                        <span class=kd>func</span>
                                        <span class=p>(</span>
                                        <span class=nx>ctx</span>
                                        <span class=w></span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>args</span>
                                        <span class=w></span>
                                        <span class=kt>string</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>opts</span>
                                        <span class=w></span>
                                        <span class=o>...</span>
                                        <span class=nx>tool</span>
                                        <span class=p>.</span>
                                        <span class=nx>Option</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>(</span>
                                        <span class=kt>string</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=kt>error</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>result</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>err</span>
                                        <span class=w></span>
                                        <span class=o>:=</span>
                                        <span class=w></span>
                                        <span class=nf>endpoint</span>
                                        <span class=p>(</span>
                                        <span class=nx>ctx</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>args</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>opts</span>
                                        <span class=o>...</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=k>if</span>
                                        <span class=w></span>
                                        <span class=nx>err</span>
                                        <span class=w></span>
                                        <span class=o>!=</span>
                                        <span class=w></span>
                                        <span class=kc>nil</span>
                                        <span class=w></span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// 将错误转换为字符串，而不是返回错误</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=k>return</span>
                                        <span class=w></span>
                                        <span class=nx>fmt</span>
                                        <span class=p>.</span>
                                        <span class=nf>Sprintf</span>
                                        <span class=p>(</span>
                                        <span class=s>&#34;[tool error] %v &#34;</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>err</span>
                                        <span class=p>),</span>
                                        <span class=w></span>
                                        <span class=kc>nil</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>}</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=k>return</span>
                                        <span class=w></span>
                                        <span class=nx>result</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=kc>nil</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>},</span>
                                        <span class=w></span>
                                        <span class=kc>nil</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>}</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                            </code>
                        </pre>
                    </div>
                    <p>
                        <strong>效果：</strong>
                    </p>
                    <pre tabindex=0>
                        <code>[tool call] read_file(file_path: &#34;nonexistent.txt &#34;)
[tool result] [tool error] open nonexistent.txt: no such file or directory
[assistant] 抱歉，文件不存在，请检查文件路径...
// 对话继续，模型可以根据错误信息调整策略
</code>
                    </pre>
                    <h3 id=modelretryconfig>ModelRetryConfig</h3>
                    <p>
                        <code>ModelRetryConfig</code>
                        配置 ChatModel 的自动重试：
                    </p>
                    <div class=highlight>
                        <pre tabindex=0 class=chroma>
                            <code class=language-go data-lang=go>
                                <span class=line>
                                    <span class=cl>
                                        <span class=kd>type</span>
                                        <span class=w></span>
                                        <span class=nx>ModelRetryConfig</span>
                                        <span class=w></span>
                                        <span class=kd>struct</span>
                                        <span class=w></span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>MaxRetries</span>
                                        <span class=w></span>
                                        <span class=kt>int</span>
                                        <span class=w></span>
                                        <span class=c1>// 最大重试次数</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>IsRetryAble</span>
                                        <span class=w></span>
                                        <span class=kd>func</span>
                                        <span class=p>(</span>
                                        <span class=nx>ctx</span>
                                        <span class=w></span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>err</span>
                                        <span class=w></span>
                                        <span class=kt>error</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=kt>bool</span>
                                        <span class=w></span>
                                        <span class=c1>// 判断是否可重试</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>}</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                            </code>
                        </pre>
                    </div>
                    <p>
                        <strong>使用方式（以 DeepAgent 为例）：</strong>
                    </p>
                    <div class=highlight>
                        <pre tabindex=0 class=chroma>
                            <code class=language-go data-lang=go>
                                <span class=line>
                                    <span class=cl>
                                        <span class=nx>agent</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>err</span>
                                        <span class=w></span>
                                        <span class=o>:=</span>
                                        <span class=w></span>
                                        <span class=nx>deep</span>
                                        <span class=p>.</span>
                                        <span class=nf>New</span>
                                        <span class=p>(</span>
                                        <span class=nx>ctx</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=o>&amp;</span>
                                        <span class=nx>deep</span>
                                        <span class=p>.</span>
                                        <span class=nx>Config</span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// ...</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>ModelRetryConfig</span>
                                        <span class=p>:</span>
                                        <span class=w></span>
                                        <span class=o>&amp;</span>
                                        <span class=nx>adk</span>
                                        <span class=p>.</span>
                                        <span class=nx>ModelRetryConfig</span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>MaxRetries</span>
                                        <span class=p>:</span>
                                        <span class=w></span>
                                        <span class=mi>5</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>IsRetryAble</span>
                                        <span class=p>:</span>
                                        <span class=w></span>
                                        <span class=kd>func</span>
                                        <span class=p>(</span>
                                        <span class=nx>_</span>
                                        <span class=w></span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>err</span>
                                        <span class=w></span>
                                        <span class=kt>error</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=kt>bool</span>
                                        <span class=w></span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// 429 限流错误可重试</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=k>return</span>
                                        <span class=w></span>
                                        <span class=nx>strings</span>
                                        <span class=p>.</span>
                                        <span class=nf>Contains</span>
                                        <span class=p>(</span>
                                        <span class=nx>err</span>
                                        <span class=p>.</span>
                                        <span class=nf>Error</span>
                                        <span class=p>(),</span>
                                        <span class=w></span>
                                        <span class=s>&#34;429 &#34;</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=o>||</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>strings</span>
                                        <span class=p>.</span>
                                        <span class=nf>Contains</span>
                                        <span class=p>(</span>
                                        <span class=nx>err</span>
                                        <span class=p>.</span>
                                        <span class=nf>Error</span>
                                        <span class=p>(),</span>
                                        <span class=w></span>
                                        <span class=s>&#34;Too Many Requests &#34;</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=o>||</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>strings</span>
                                        <span class=p>.</span>
                                        <span class=nf>Contains</span>
                                        <span class=p>(</span>
                                        <span class=nx>err</span>
                                        <span class=p>.</span>
                                        <span class=nf>Error</span>
                                        <span class=p>(),</span>
                                        <span class=w></span>
                                        <span class=s>&#34;qpm limit &#34;</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>},</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>},</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>})</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                            </code>
                        </pre>
                    </div>
                    <p>
                        <strong>重试策略：</strong>
                    </p>
                    <ul>
                        <li>指数退避：每次重试间隔递增</li>
                        <li>
                            可配置条件：通过 <code>IsRetryAble</code>
                            判断哪些错误可重试
                        </li>
                        <li>自动恢复：无需用户干预</li>
                    </ul>
                    <h2 id=middleware-的实现>Middleware 的实现</h2>
                    <h3 id=1-实现-safetoolmiddleware>1. 实现 SafeToolMiddleware</h3>
                    <div class=highlight>
                        <pre tabindex=0 class=chroma>
                            <code class=language-go data-lang=go>
                                <span class=line>
                                    <span class=cl>
                                        <span class=kd>type</span>
                                        <span class=w></span>
                                        <span class=nx>safeToolMiddleware</span>
                                        <span class=w></span>
                                        <span class=kd>struct</span>
                                        <span class=w></span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>adk</span>
                                        <span class=p>.</span>
                                        <span class=nx>BaseChatModelAgentMiddleware</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>}</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=kd>func</span>
                                        <span class=w></span>
                                        <span class=p>(</span>
                                        <span class=nx>m</span>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>safeToolMiddleware</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=nf>WrapInvokableToolCall</span>
                                        <span class=p>(</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>_</span>
                                        <span class=w></span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>endpoint</span>
                                        <span class=w></span>
                                        <span class=nx>adk</span>
                                        <span class=p>.</span>
                                        <span class=nx>InvokableToolCallEndpoint</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>_</span>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>adk</span>
                                        <span class=p>.</span>
                                        <span class=nx>ToolContext</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>(</span>
                                        <span class=nx>adk</span>
                                        <span class=p>.</span>
                                        <span class=nx>InvokableToolCallEndpoint</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=kt>error</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=k>return</span>
                                        <span class=w></span>
                                        <span class=kd>func</span>
                                        <span class=p>(</span>
                                        <span class=nx>ctx</span>
                                        <span class=w></span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>args</span>
                                        <span class=w></span>
                                        <span class=kt>string</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>opts</span>
                                        <span class=w></span>
                                        <span class=o>...</span>
                                        <span class=nx>tool</span>
                                        <span class=p>.</span>
                                        <span class=nx>Option</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>(</span>
                                        <span class=kt>string</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=kt>error</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>result</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>err</span>
                                        <span class=w></span>
                                        <span class=o>:=</span>
                                        <span class=w></span>
                                        <span class=nf>endpoint</span>
                                        <span class=p>(</span>
                                        <span class=nx>ctx</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>args</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>opts</span>
                                        <span class=o>...</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=k>if</span>
                                        <span class=w></span>
                                        <span class=nx>err</span>
                                        <span class=w></span>
                                        <span class=o>!=</span>
                                        <span class=w></span>
                                        <span class=kc>nil</span>
                                        <span class=w></span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// 中断错误不转换，需要继续传播</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=k>if</span>
                                        <span class=w></span>
                                        <span class=nx>_</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>ok</span>
                                        <span class=w></span>
                                        <span class=o>:=</span>
                                        <span class=w></span>
                                        <span class=nx>compose</span>
                                        <span class=p>.</span>
                                        <span class=nf>IsInterruptRerunError</span>
                                        <span class=p>(</span>
                                        <span class=nx>err</span>
                                        <span class=p>);</span>
                                        <span class=w></span>
                                        <span class=nx>ok</span>
                                        <span class=w></span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=k>return</span>
                                        <span class=w></span>
                                        <span class=s>&#34;&#34;</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>err</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>}</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// 其他错误转换为字符串</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=k>return</span>
                                        <span class=w></span>
                                        <span class=nx>fmt</span>
                                        <span class=p>.</span>
                                        <span class=nf>Sprintf</span>
                                        <span class=p>(</span>
                                        <span class=s>&#34;[tool error] %v &#34;</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>err</span>
                                        <span class=p>),</span>
                                        <span class=w></span>
                                        <span class=kc>nil</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>}</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=k>return</span>
                                        <span class=w></span>
                                        <span class=nx>result</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=kc>nil</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>},</span>
                                        <span class=w></span>
                                        <span class=kc>nil</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>}</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                            </code>
                        </pre>
                    </div>
                    <h3 id=2-实现流式-tool-错误处理>2. 实现流式 Tool 错误处理</h3>
                    <div class=highlight>
                        <pre tabindex=0 class=chroma>
                            <code class=language-go data-lang=go>
                                <span class=line>
                                    <span class=cl>
                                        <span class=kd>func</span>
                                        <span class=w></span>
                                        <span class=p>(</span>
                                        <span class=nx>m</span>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>safeToolMiddleware</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=nf>WrapStreamableToolCall</span>
                                        <span class=p>(</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>_</span>
                                        <span class=w></span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>endpoint</span>
                                        <span class=w></span>
                                        <span class=nx>adk</span>
                                        <span class=p>.</span>
                                        <span class=nx>StreamableToolCallEndpoint</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>_</span>
                                        <span class=w></span>
                                        <span class=o>*</span>
                                        <span class=nx>adk</span>
                                        <span class=p>.</span>
                                        <span class=nx>ToolContext</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>(</span>
                                        <span class=nx>adk</span>
                                        <span class=p>.</span>
                                        <span class=nx>StreamableToolCallEndpoint</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=kt>error</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=k>return</span>
                                        <span class=w></span>
                                        <span class=kd>func</span>
                                        <span class=p>(</span>
                                        <span class=nx>ctx</span>
                                        <span class=w></span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>args</span>
                                        <span class=w></span>
                                        <span class=kt>string</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>opts</span>
                                        <span class=w></span>
                                        <span class=o>...</span>
                                        <span class=nx>tool</span>
                                        <span class=p>.</span>
                                        <span class=nx>Option</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>(</span>
                                        <span class=o>*</span>
                                        <span class=nx>schema</span>
                                        <span class=p>.</span>
                                        <span class=nx>StreamReader</span>
                                        <span class=p>[</span>
                                        <span class=kt>string</span>
                                        <span class=p>],</span>
                                        <span class=w></span>
                                        <span class=kt>error</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>sr</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>err</span>
                                        <span class=w></span>
                                        <span class=o>:=</span>
                                        <span class=w></span>
                                        <span class=nf>endpoint</span>
                                        <span class=p>(</span>
                                        <span class=nx>ctx</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>args</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>opts</span>
                                        <span class=o>...</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=k>if</span>
                                        <span class=w></span>
                                        <span class=nx>err</span>
                                        <span class=w></span>
                                        <span class=o>!=</span>
                                        <span class=w></span>
                                        <span class=kc>nil</span>
                                        <span class=w></span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=k>if</span>
                                        <span class=w></span>
                                        <span class=nx>_</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>ok</span>
                                        <span class=w></span>
                                        <span class=o>:=</span>
                                        <span class=w></span>
                                        <span class=nx>compose</span>
                                        <span class=p>.</span>
                                        <span class=nf>IsInterruptRerunError</span>
                                        <span class=p>(</span>
                                        <span class=nx>err</span>
                                        <span class=p>);</span>
                                        <span class=w></span>
                                        <span class=nx>ok</span>
                                        <span class=w></span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=k>return</span>
                                        <span class=w></span>
                                        <span class=kc>nil</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>err</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>}</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// 返回包含错误信息的单帧流</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=k>return</span>
                                        <span class=w></span>
                                        <span class=nf>singleChunkReader</span>
                                        <span class=p>(</span>
                                        <span class=nx>fmt</span>
                                        <span class=p>.</span>
                                        <span class=nf>Sprintf</span>
                                        <span class=p>(</span>
                                        <span class=s>&#34;[tool error] %v &#34;</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>err</span>
                                        <span class=p>)),</span>
                                        <span class=w></span>
                                        <span class=kc>nil</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>}</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=c1>// 包装流，捕获流中的错误</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=k>return</span>
                                        <span class=w></span>
                                        <span class=nf>safeWrapReader</span>
                                        <span class=p>(</span>
                                        <span class=nx>sr</span>
                                        <span class=p>),</span>
                                        <span class=w></span>
                                        <span class=kc>nil</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>},</span>
                                        <span class=w></span>
                                        <span class=kc>nil</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>}</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                            </code>
                        </pre>
                    </div>
                    <h3 id=3-配置-agent-使用-middleware>3. 配置 Agent 使用 Middleware</h3>
                    <p>
                        本章继续使用第四章引入的 <code>DeepAgent</code>
                        ，在其 <code>Handlers</code>
                        字段中注册 Middleware：
                    </p>
                    <div class=highlight>
                        <pre tabindex=0 class=chroma>
                            <code class=language-go data-lang=go>
                                <span class=line>
                                    <span class=cl>
                                        <span class=nx>agent</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>err</span>
                                        <span class=w></span>
                                        <span class=o>:=</span>
                                        <span class=w></span>
                                        <span class=nx>deep</span>
                                        <span class=p>.</span>
                                        <span class=nf>New</span>
                                        <span class=p>(</span>
                                        <span class=nx>ctx</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=o>&amp;</span>
                                        <span class=nx>deep</span>
                                        <span class=p>.</span>
                                        <span class=nx>Config</span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>Name</span>
                                        <span class=p>:</span>
                                        <span class=w></span>
                                        <span class=s>&#34;Ch05MiddlewareAgent &#34;</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>Description</span>
                                        <span class=p>:</span>
                                        <span class=w></span>
                                        <span class=s>&#34;ChatWithDoc agent with safe tool middleware and retry.&#34;</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>ChatModel</span>
                                        <span class=p>:</span>
                                        <span class=w></span>
                                        <span class=nx>cm</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>Instruction</span>
                                        <span class=p>:</span>
                                        <span class=w></span>
                                        <span class=nx>agentInstruction</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>Backend</span>
                                        <span class=p>:</span>
                                        <span class=w></span>
                                        <span class=nx>backend</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>StreamingShell</span>
                                        <span class=p>:</span>
                                        <span class=w></span>
                                        <span class=nx>backend</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>MaxIteration</span>
                                        <span class=p>:</span>
                                        <span class=w></span>
                                        <span class=mi>50</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>Handlers</span>
                                        <span class=p>:</span>
                                        <span class=w></span>
                                        <span class=p>[]</span>
                                        <span class=nx>adk</span>
                                        <span class=p>.</span>
                                        <span class=nx>ChatModelAgentMiddleware</span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=o>&amp;</span>
                                        <span class=nx>safeToolMiddleware</span>
                                        <span class=p>{},</span>
                                        <span class=w></span>
                                        <span class=c1>// 将 Tool 错误转换为字符串</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>},</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>ModelRetryConfig</span>
                                        <span class=p>:</span>
                                        <span class=w></span>
                                        <span class=o>&amp;</span>
                                        <span class=nx>adk</span>
                                        <span class=p>.</span>
                                        <span class=nx>ModelRetryConfig</span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>MaxRetries</span>
                                        <span class=p>:</span>
                                        <span class=w></span>
                                        <span class=mi>5</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>IsRetryAble</span>
                                        <span class=p>:</span>
                                        <span class=w></span>
                                        <span class=kd>func</span>
                                        <span class=p>(</span>
                                        <span class=nx>_</span>
                                        <span class=w></span>
                                        <span class=nx>context</span>
                                        <span class=p>.</span>
                                        <span class=nx>Context</span>
                                        <span class=p>,</span>
                                        <span class=w></span>
                                        <span class=nx>err</span>
                                        <span class=w></span>
                                        <span class=kt>error</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=kt>bool</span>
                                        <span class=w></span>
                                        <span class=p>{</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=k>return</span>
                                        <span class=w></span>
                                        <span class=nx>strings</span>
                                        <span class=p>.</span>
                                        <span class=nf>Contains</span>
                                        <span class=p>(</span>
                                        <span class=nx>err</span>
                                        <span class=p>.</span>
                                        <span class=nf>Error</span>
                                        <span class=p>(),</span>
                                        <span class=w></span>
                                        <span class=s>&#34;429 &#34;</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                        <span class=o>||</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=nx>strings</span>
                                        <span class=p>.</span>
                                        <span class=nf>Contains</span>
                                        <span class=p>(</span>
                                        <span class=nx>err</span>
                                        <span class=p>.</span>
                                        <span class=nf>Error</span>
                                        <span class=p>(),</span>
                                        <span class=w></span>
                                        <span class=s>&#34;Too Many Requests &#34;</span>
                                        <span class=p>)</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>},</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>},</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                                <span class=line>
                                    <span class=cl>
                                        <span class=w></span>
                                        <span class=p>})</span>
                                        <span class=w></span>
                                    </span>
                                </span>
                            </code>
                        </pre>
                    </div>
                    <p>
                        <strong>注意</strong>
                        ：<code>Handlers</code>
                        字段（在配置中）和 &ldquo;Middleware &rdquo;（在文档中讨论的概念）是同一回事——<code>Handlers</code>
                        是配置字段名，而 <code>ChatModelAgentMiddleware</code>
                        是接口名。
                    </p>
                    <pre tabindex=0>
                        <code>**关键代码片段（**注意：这是简化后的代码片段，不能直接运行，完整代码请参考** [cmd/ch05/main.go](https://github.com/cloudwego/eino-examples/blob/main/quickstart/chatwitheino/cmd/ch05/main.go)）：
`
	cleaned := PreCleanHTML(html)
	t.Logf("cleaned: %s", cleaned)
}
