.PHONY: help dev build test tidy clean fmt vet lint \
        release create-tags delete-tags push-tags \
        create-remote-tags delete-remote-tags list-tags \
        init-work sync-work update-deps vuln-check \
        ai-check ai-audit ai-verify ai-quick godoc-coverage \
        ai-docs-sync ai-deps-check skill-sync skill-check \
        doc-update \
        deps-visualize deps-mermaid deps-json \
        changelog-check changelog-update \
        quality-score quality-report \
        _backup-gowork _restore-gowork \
        _phase1-main _commit-main _create-main-tag _push-main-tag \
        _phase2-sub _update-sub-deps _tidy-sub-modules _commit-sub-modules _create-sub-tags _push-sub-tags \
        _phase3-examples _update-example-deps _tidy-example-modules _commit-examples \
        _push-code push-code _validate-version

# ============================================================================
# Variables
# ============================================================================

DEFAULT_VERSION := v0.0.0
GO := go
REMOTE := origin
BRANCH := master
MAIN_MODULE := github.com/xudefa/enhance
VERSION := $(or $(VERSION),$(DEFAULT_VERSION))
TAG_LIST_LIMIT := 30
SED_INPLACE := sed -i ''
COLOR_RED := \033[0;31m
COLOR_GREEN := \033[0;32m
COLOR_YELLOW := \033[1;33m
COLOR_CYAN := \033[0;36m
COLOR_RESET := \033[0m

ALL_MODULES := . $(SUB_MODULES)

SUB_MODULES := $(shell find . -name "go.mod" -not -path "./go.mod" -exec dirname {} \; | sort)

STARTER_MODULES := $(shell find ./starter -name "go.mod" -exec dirname {} \; | sort 2>/dev/null)

EXAMPLE_MODULES := $(shell find ./examples -name "go.mod" -exec dirname {} \; | sort 2>/dev/null)

define run-all-modules
@failed=""; \
for dir in $(ALL_MODULES); do \
	echo "[$(1)] $$dir"; \
	$(GO) -C "$$dir" $(1) ./... 2>&1 || { echo "  [失败] $$dir"; failed="$$failed $$dir"; }; \
done; \
if [ -n "$$failed" ]; then echo "$(COLOR_RED)❌ $(1) 失败:$$failed$(COLOR_RESET)"; exit 1; fi
endef

define git-commit-if-changed
@git diff --cached --quiet && echo "$(COLOR_YELLOW)⚠️  [跳过] $(1) 无变更$(COLOR_RESET)" || git commit -m "$(2)"
endef

# ============================================================================
# Help
# ============================================================================

help: ## 显示帮助信息
	@echo "Enhance 多模块管理工具"
	@echo ""
	@echo "用法: make [target] [VERSION=vX.Y.Z] [BRANCH=master]"
	@echo ""
	@echo "开发模式:"
	@echo "  dev             - 进入开发模式（go.work 链接所有模块）"
	@echo "  build           - 构建所有模块"
	@echo "  test            - 运行所有测试"
	@echo "  tidy            - 批量 go mod tidy"
	@echo "  fmt             - 格式化所有代码"
	@echo "  vet             - 静态分析所有代码"
	@echo "  lint            - 综合检查（fmt + vet + build）"
	@echo "  clean           - 清理构建产物"
	@echo ""
	@echo "AI 可维护性检查:"
	@echo "  ai-check        - 运行 AI 可维护性检查"
	@echo "  godoc-coverage  - 检查 godoc 覆盖率"
	@echo "  ai-audit        - 生成 AI 可读性审计报告"
	@echo "  ai-verify       - 运行完整 AI 验证（检查 + 审计）"
	@echo "  ai-quick        - 运行快速 AI 检查（lint + test）"
	@echo "  ai-docs-sync    - 检查文档与代码是否同步"
	@echo "  ai-deps-check   - 检查依赖方向是否正确"
	@echo ""
	@echo "文档管理:"
	@echo "  doc-update      - 更新文档（基于代码生成）"
	@echo ""
	@echo "依赖可视化:"
	@echo "  deps-visualize  - 生成包间依赖可视化"
	@echo "  deps-mermaid    - 生成 Mermaid 格式依赖图"
	@echo "  deps-json       - 生成 JSON 格式依赖数据"
	@echo ""
	@echo "API 变更日志:"
	@echo "  changelog-check - 检查 API 变更是否记录"
	@echo "  changelog-update - 更新 CHANGELOG.md"
	@echo ""
	@echo "代码质量评分:"
	@echo "  quality-score   - 运行代码质量评分"
	@echo "  quality-report  - 生成代码质量报告"
	@echo ""
	@echo "版本发布:"
	@echo "  release         - 完整发布流程（replace + tidy + tag + push）"
	@echo "  create-tags     - 创建本地 git tags"
	@echo "  push-tags       - 推送 tags 到远端"
	@echo "  delete-tags     - 删除本地 tags"
	@echo "  list-tags       - 列出所有 tags"
	@echo ""
	@echo "状态检查:"
	@echo "  status          - 显示项目状态"
	@echo ""
	@echo "示例:"
	@echo "  make dev                           # 进入开发模式"
	@echo "  make ai-verify                     # 运行完整 AI 验证"
	@echo "  make quality-score                 # 运行代码质量评分"
	@echo "  make changelog-check               # 检查 API 变更记录"
	@echo "  make release VERSION=v0.1.0        # 发布 v0.1.0"
	@echo "  make release VERSION=v0.1.0 BRANCH=main  # 发布到 main 分支"
	@echo "  make create-tags VERSION=v0.1.0    # 仅创建 tags"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ============================================================================
# Development Mode
# ============================================================================

dev: ## 进入开发模式（go.work 链接所有模块）
	@echo "=== 开发模式 ==="
	@if [ ! -f "go.work" ]; then \
		echo "$(COLOR_RED)[错误] 未找到 go.work，请先运行 make init-work$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "同步 go.work..."
	@$(GO) work sync
	@echo ""
	@echo "$(COLOR_GREEN)✅ 开发模式就绪$(COLOR_RESET)"
	@echo "   go.work 已链接 $(words $(SUB_MODULES)) 个子模块 + 根模块"
	@echo "   所有模块通过 workspace 解析依赖，无需 replace"

init-work: ## 初始化 go.work 文件
	@echo "=== 初始化 go.work ==="
	@if [ -f "go.work" ]; then \
		echo "go.work 已存在，跳过"; \
		exit 0; \
	fi
	@$(GO) work init
	@find . -name "go.mod" -not -path "./go.mod" -exec sh -c '$(GO) work use "$$(dirname "$$1")"' _ {} \;
	@echo "$(COLOR_GREEN)✅ go.work 已创建$(COLOR_RESET)"

sync-work: ## 同步 go.work 依赖
	@$(GO) work sync
	@echo "$(COLOR_GREEN)✅ 依赖已同步$(COLOR_RESET)"

# ============================================================================
# Build & Test
# ============================================================================

build: ## 构建所有模块
	@echo "=== 构建所有模块 ==="
	$(call run-all-modules,build)
	@echo "$(COLOR_GREEN)✅ 构建完成$(COLOR_RESET)"

test: ## 运行所有测试
	@echo "=== 运行所有测试 ==="
	$(call run-all-modules,test)
	@echo "$(COLOR_GREEN)✅ 测试完成$(COLOR_RESET)"

tidy: ## 批量 go mod tidy
	@echo "=== go mod tidy ==="
	@failed=""; \
	for dir in $(ALL_MODULES); do \
		echo "[tidy] $$dir"; \
		$(GO) -C "$$dir" mod tidy 2>&1 || { echo "  [失败] $$dir"; failed="$$failed $$dir"; }; \
	done; \
	if [ -n "$$failed" ]; then echo "$(COLOR_RED)❌ tidy 失败:$$failed$(COLOR_RESET)"; exit 1; fi
	@echo "$(COLOR_GREEN)✅ tidy 完成$(COLOR_RESET)"

clean: ## 清理构建产物
	@find . -name "*.test" -type f -delete
	@find . -name "*.out" -type f -delete
	@rm -rf coverage.out
	@rm -f go.work.back
	@echo "$(COLOR_GREEN)✅ 清理完成$(COLOR_RESET)"

fmt: ## 格式化所有代码
	@echo "=== 格式化代码 ==="
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "以下文件需要格式化:"; \
		echo "$$unformatted"; \
		gofmt -w .; \
		echo "$(COLOR_GREEN)✅ 格式化完成$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_GREEN)✅ 代码已格式化$(COLOR_RESET)"; \
	fi

vet: ## 静态分析所有代码
	@echo "=== 静态分析 ==="
	$(call run-all-modules,vet)
	@echo "$(COLOR_GREEN)✅ 静态分析完成$(COLOR_RESET)"

lint: fmt vet build ## 综合检查（fmt + vet + build）

# ============================================================================
# Release
# ============================================================================

release: _validate-version _backup-gowork _phase1-main _phase2-sub _phase3-examples _push-code _restore-gowork ## 完整发布流程（三阶段提交）
	@echo ""
	@echo "$(COLOR_GREEN)✅ 发布完成: $(VERSION)$(COLOR_RESET)"
	@echo ""
	@echo "发布内容:"
	@echo "  - 主模块 tag: $(VERSION)"
	@echo "  - 子模块 tags: starter/*/$(VERSION)"
	@echo "  - examples 依赖已更新"
	@echo "  - 代码已推送到 $(REMOTE)/$(BRANCH)"
	@echo "  - go.work 已恢复"

_validate-version: ## [内部] 验证版本号格式
	@if ! echo "$(VERSION)" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?(\+[a-zA-Z0-9.]+)?$$'; then \
		echo "$(COLOR_RED)❌ [错误] 版本号格式无效: $(VERSION)$(COLOR_RESET)"; \
		echo "期望格式: vX.Y.Z 或 vX.Y.Z-rc1 或 vX.Y.Z+build"; \
		exit 1; \
	fi
	@if [ "$(VERSION)" = "$(DEFAULT_VERSION)" ]; then \
		echo "$(COLOR_YELLOW)⚠️  警告: 使用默认版本号 $(DEFAULT_VERSION)，请确认是否需要指定 VERSION=vX.Y.Z$(COLOR_RESET)"; \
	fi

# ============================================================================
# Go Work Backup & Restore
# ============================================================================

_backup-gowork: ## [内部] 备份 go.work 文件
	@echo "=== 备份 go.work ==="
	@if [ -f "go.work" ]; then \
		cp go.work go.work.back && echo "  [备份] go.work -> go.work.back"; \
	else \
		echo "  [跳过] go.work 不存在"; \
	fi

_restore-gowork: ## [内部] 恢复 go.work 文件
	@echo "=== 恢复 go.work ==="
	@if [ -f "go.work.back" ]; then \
		mv go.work.back go.work && echo "  [恢复] go.work.back -> go.work"; \
	else \
		echo "  [跳过] 无备份文件"; \
	fi

# ============================================================================
# Phase 1: 主模块发布
# ============================================================================

_phase1-main: _commit-main _create-main-tag _push-main-tag ## 阶段1：提交主模块代码并打tag
	@echo ""
	@echo "$(COLOR_GREEN)✅ 阶段1完成: 主模块 $(VERSION) 已发布$(COLOR_RESET)"

_commit-main: ## [内部] 提交主模块代码
	@echo "=== 阶段1: 提交主模块代码 ==="
	@git add go.mod
	@test -f go.sum && git add go.sum || true
	@find . -maxdepth 1 -name "*.go" -exec git add {} +
	@test -f go.work && git add go.work || true
	@test -f go.work.sum && git add go.work.sum || true
	$(call git-commit-if-changed,主模块,release: main module $(VERSION))

_create-main-tag: ## [内部] 创建主模块 tag
	@echo "=== 创建主模块 tag: $(VERSION) ==="
	@if git rev-parse "$(VERSION)" >/dev/null 2>&1; then \
		echo "  [跳过] 已存在"; \
	else \
		git tag -a "$(VERSION)" -m "Release $(VERSION)" && echo "  [创建] ✅"; \
	fi

_push-main-tag: ## [内部] 推送主模块 tag
	@echo "=== 推送主模块 tag: $(VERSION) ==="
	@git push $(REMOTE) "$(VERSION)" && echo "  [推送] ✅" || { \
		echo "$(COLOR_RED)❌ [错误] 推送主模块 tag 失败，流程已中断$(COLOR_RESET)"; \
		echo "请手动执行: git push $(REMOTE) $(VERSION)"; \
		exit 1; \
	}

# ============================================================================
# Phase 2: 子模块发布
# ============================================================================

_phase2-sub: _update-sub-deps _tidy-sub-modules _commit-sub-modules _create-sub-tags _push-sub-tags ## 阶段2：更新子模块依赖并发布
	@echo ""
	@echo "$(COLOR_GREEN)✅ 阶段2完成: 子模块 tags 已发布$(COLOR_RESET)"

_update-sub-deps: ## [内部] 更新子模块依赖到主模块新版本
	@echo "=== 阶段2: 更新子模块依赖到 $(VERSION) ==="
	@for dir in $(STARTER_MODULES); do \
		modfile="$$dir/go.mod"; \
		if [ ! -f "$$modfile" ]; then continue; fi; \
		if grep -q "$(MAIN_MODULE) " "$$modfile"; then \
			$(SED_INPLACE) 's|$(MAIN_MODULE) v[0-9]\+\.[0-9]\+\.[0-9]\+\([^/]*\)|$(MAIN_MODULE) $(VERSION)\1|g' "$$modfile"; \
			echo "  [更新] $$dir -> $(VERSION)"; \
		fi; \
	done

_tidy-sub-modules: ## [内部] 子模块 go mod tidy
	@echo "=== 子模块 go mod tidy ==="
	@failed=""; \
	for dir in $(STARTER_MODULES); do \
		echo "[tidy] $$dir"; \
		$(GO) -C "$$dir" mod tidy 2>&1 || { echo "  [失败] $$dir"; failed="$$failed $$dir"; }; \
	done; \
	if [ -n "$$failed" ]; then echo "$(COLOR_RED)❌ 子模块 tidy 失败:$$failed$(COLOR_RESET)"; exit 1; fi

_commit-sub-modules: ## [内部] 提交子模块代码
	@echo "=== 提交子模块代码 ==="
	@git add starter/*/go.mod starter/*/go.sum 2>/dev/null || true
	$(call git-commit-if-changed,子模块,release: sub modules depend on main $(VERSION))

_create-sub-tags: ## [内部] 创建子模块 tags
	@echo "=== 创建子模块 tags ==="
	@for dir in $(STARTER_MODULES); do \
		tag="$${dir#./}/$(VERSION)"; \
		echo "  [tag] $$tag"; \
		if git rev-parse "$$tag" >/dev/null 2>&1; then \
			echo "    [跳过] 已存在"; \
		else \
			git tag -a "$$tag" -m "Release $$tag" && echo "    [创建] ✅"; \
		fi; \
	done

_push-sub-tags: ## [内部] 推送子模块 tags
	@echo "=== 推送子模块 tags ==="
	@failed=""; \
	for dir in $(STARTER_MODULES); do \
		tag="$${dir#./}/$(VERSION)"; \
		if git rev-parse "$$tag" >/dev/null 2>&1; then \
			git push $(REMOTE) "$$tag" 2>&1 && echo "  [推送] $$tag ✅" || { echo "  [失败] $$tag"; failed="$$failed $$tag"; }; \
		fi; \
	done; \
	if [ -n "$$failed" ]; then \
		echo "$(COLOR_RED)❌ [错误] 以下子模块 tags 推送失败:$$failed$(COLOR_RESET)"; \
		echo "流程已中断，请手动执行:"; \
		for tag in $$failed; do echo "  git push $(REMOTE) $$tag"; done; \
		exit 1; \
	fi

# ============================================================================
# Phase 3: Examples 更新
# ============================================================================

_phase3-examples: _update-example-deps _tidy-example-modules _commit-examples ## 阶段3：更新 examples 依赖并提交
	@echo ""
	@echo "$(COLOR_GREEN)✅ 阶段3完成: examples 依赖已更新$(COLOR_RESET)"

_update-example-deps: ## [内部] 更新 examples 中的 starter 依赖版本
	@echo "=== 阶段3: 更新 examples 依赖到 $(VERSION) ==="
	@for dir in $(EXAMPLE_MODULES); do \
		modfile="$$dir/go.mod"; \
		if [ ! -f "$$modfile" ]; then continue; fi; \
		updated=false; \
		for starter in $(STARTER_MODULES); do \
			starter_mod="$$starter/go.mod"; \
			if [ ! -f "$$starter_mod" ]; then continue; fi; \
			starter_path=$$(grep "^module " "$$starter_mod" | awk '{print $$2}'); \
			if [ -z "$$starter_path" ]; then continue; fi; \
			if grep -q "$$starter_path " "$$modfile"; then \
				$(SED_INPLACE) "s|$$starter_path v[0-9]\+\.[0-9]\+\.[0-9]\+\([^/]*\)|$$starter_path $(VERSION)\1|g" "$$modfile"; \
				updated=true; \
			fi; \
		done; \
		if [ "$$updated" = true ]; then echo "  [更新] $$dir"; fi; \
	done

_tidy-example-modules: ## [内部] examples go mod tidy
	@echo "=== examples go mod tidy ==="
	@failed=""; \
	for dir in $(EXAMPLE_MODULES); do \
		echo "[tidy] $$dir"; \
		$(GO) -C "$$dir" mod tidy 2>&1 || { echo "  [失败] $$dir"; failed="$$failed $$dir"; }; \
	done; \
	if [ -n "$$failed" ]; then echo "$(COLOR_RED)❌ examples tidy 失败:$$failed$(COLOR_RESET)"; exit 1; fi

_commit-examples: ## [内部] 提交 examples 变更
	@echo "=== 提交 examples 变更 ==="
	@git add examples/*/go.mod examples/*/go.sum 2>/dev/null || true
	$(call git-commit-if-changed,examples,release: update examples dependencies to $(VERSION))

# ============================================================================
# Tag Management
# ============================================================================

create-tags: _validate-version ## 创建本地 git tags
	@echo "=== 创建本地 tags (VERSION=$(VERSION)) ==="
	@echo ""
	@echo "主模块: $(VERSION)"
	@if git rev-parse "$(VERSION)" >/dev/null 2>&1; then \
		echo "  [跳过] 已存在"; \
	else \
		git tag -a "$(VERSION)" -m "Release $(VERSION)" && echo "  [创建] ✅"; \
	fi
	@echo ""
	@for dir in $(STARTER_MODULES); do \
		tag="$${dir#./}/$(VERSION)"; \
		echo "子模块: $$tag"; \
		if git rev-parse "$$tag" >/dev/null 2>&1; then \
			echo "  [跳过] 已存在"; \
		else \
			git tag -a "$$tag" -m "Release $$tag" && echo "  [创建] ✅"; \
		fi; \
	done
	@echo ""
	@echo "$(COLOR_GREEN)✅ 本地 tags 创建完成$(COLOR_RESET)"

delete-tags: ## 删除本地 git tags
	@echo "=== 删除本地 tags (VERSION=$(VERSION)) ==="
	@if git rev-parse "$(VERSION)" >/dev/null 2>&1; then \
		git tag -d "$(VERSION)" && echo "  [删除] $(VERSION) ✅"; \
	else \
		echo "  [跳过] $(VERSION) 不存在"; \
	fi
	@for dir in $(STARTER_MODULES); do \
		tag="$${dir#./}/$(VERSION)"; \
		if git rev-parse "$$tag" >/dev/null 2>&1; then \
			git tag -d "$$tag" && echo "  [删除] $$tag ✅"; \
		fi; \
	done
	@echo "$(COLOR_GREEN)✅ 本地 tags 删除完成$(COLOR_RESET)"

push-tags: ## 推送 tags 到远端
	@echo "=== 推送 tags 到 $(REMOTE) ==="
	@failed=""; \
	if git rev-parse "$(VERSION)" >/dev/null 2>&1; then \
		git push $(REMOTE) "$(VERSION)" && echo "  [推送] $(VERSION) ✅" || { echo "  [失败] $(VERSION)"; failed="$(VERSION)"; }; \
	fi; \
	for dir in $(STARTER_MODULES); do \
		tag="$${dir#./}/$(VERSION)"; \
		if git rev-parse "$$tag" >/dev/null 2>&1; then \
			git push $(REMOTE) "$$tag" 2>&1 && echo "  [推送] $$tag ✅" || { echo "  [失败] $$tag"; failed="$$failed $$tag"; }; \
		fi; \
	done; \
	if [ -n "$$failed" ]; then \
		echo "$(COLOR_RED)❌ [错误] 以下 tags 推送失败:$$failed$(COLOR_RESET)"; \
		echo "请手动执行:"; \
		for tag in $$failed; do echo "  git push $(REMOTE) $$tag"; done; \
		exit 1; \
	fi
	@echo "$(COLOR_GREEN)✅ Tags 推送完成$(COLOR_RESET)"

_push-code: ## [内部] 推送代码到远端
	@echo "=== 推送代码到 $(REMOTE)/$(BRANCH) ==="
	@git push $(REMOTE) $(BRANCH) && echo "$(COLOR_GREEN)✅ 代码推送完成$(COLOR_RESET)" || { \
		echo "$(COLOR_RED)❌ [错误] 推送代码失败，流程已中断$(COLOR_RESET)"; \
		echo "请手动执行: git push $(REMOTE) $(BRANCH)"; \
		exit 1; \
	}

push-code: ## 推送代码到远端
	@echo "=== 推送代码到 $(REMOTE)/$(BRANCH) ==="
	@git push $(REMOTE) $(BRANCH) && echo "$(COLOR_GREEN)✅ 代码推送完成$(COLOR_RESET)" || { \
		echo "$(COLOR_RED)❌ [错误] 推送代码失败$(COLOR_RESET)"; \
		echo "请手动执行: git push $(REMOTE) $(BRANCH)"; \
		exit 1; \
	}

create-remote-tags: create-tags push-tags ## 创建并推送 tags

delete-remote-tags: ## 删除远端 tags
	@echo "=== 删除远端 tags ($(REMOTE)) ==="
	@if git ls-remote --tags $(REMOTE) 2>/dev/null | grep -qE "refs/tags/$(VERSION)$$"; then \
		git push $(REMOTE) --delete "$(VERSION)" && echo "  [删除] $(VERSION) ✅"; \
	fi
	@for dir in $(STARTER_MODULES); do \
		tag="$${dir#./}/$(VERSION)"; \
		if git ls-remote --tags $(REMOTE) 2>/dev/null | grep -qE "refs/tags/$$tag$$"; then \
			git push $(REMOTE) --delete "$$tag" 2>/dev/null && echo "  [删除] $$tag ✅"; \
		fi; \
	done
	@echo "$(COLOR_GREEN)✅ 远端 tags 删除完成$(COLOR_RESET)"

list-tags: ## 列出所有 tags
	@echo "=== 本地 Tags ==="
	@git tag -l | sort | head -$(TAG_LIST_LIMIT)
	@echo ""
	@echo "=== 远端 Tags ($(REMOTE)) ==="
	@git ls-remote --tags $(REMOTE) 2>/dev/null | awk '{print $$2}' | sed 's|refs/tags/||' | sort | head -$(TAG_LIST_LIMIT)

# ============================================================================
# Status & Diagnostics
# ============================================================================

status: ## 显示项目状态
	@echo "=== Enhance 项目状态 ==="
	@echo ""
	@echo "📦 模块: $(words $(SUB_MODULES)) 个子模块 + 1 个根模块"
	@echo "🔧 go.work: $$(test -f go.work && echo '✅ 存在' || echo '❌ 不存在')"
	@echo "🌿 分支: $$(git branch --show-current 2>/dev/null || echo 'unknown')"
	@echo "📌 远端: $(REMOTE)"
	@echo ""
	@echo "📋 子模块列表:"
	@for dir in $(SUB_MODULES); do \
		modpath=$$(grep "^module " "$$dir/go.mod" 2>/dev/null | awk '{print $$2}'); \
		[ -z "$$modpath" ] && continue; \
		has_require=$$(grep -q "$(MAIN_MODULE)" "$$dir/go.mod" 2>/dev/null && echo "✅" || echo "❌"); \
		echo "  $$modpath (require enhance: $$has_require)"; \
	done
	@echo ""
	@echo "🏷️  Tags (VERSION=$(VERSION)):"
	@echo -n "  根模块 $(VERSION): "; \
		if git rev-parse "$(VERSION)" >/dev/null 2>&1; then echo "✅"; else echo "❌"; fi
	@echo ""
	@echo "📊 Git:"
	@git status --short 2>/dev/null | head -10

update-deps: ## 更新所有依赖
	@echo "=== 更新依赖 ==="
	@failed=""; \
	for dir in $(ALL_MODULES); do \
		echo "[update] $$dir"; \
		$(GO) -C "$$dir" get -u ./... 2>&1 | tail -2; \
		$(GO) -C "$$dir" mod tidy 2>&1 | tail -2 || { failed="$$failed $$dir"; }; \
	done; \
	if [ -n "$$failed" ]; then echo "$(COLOR_RED)❌ 依赖更新失败:$$failed$(COLOR_RESET)"; exit 1; fi
	@echo "$(COLOR_GREEN)✅ 更新完成$(COLOR_RESET)"

vuln-check: ## 检查安全漏洞
	@echo "=== 漏洞检查 ==="
	@if command -v govulncheck >/dev/null 2>&1; then \
		failed=""; \
		for dir in $(ALL_MODULES); do \
			echo "[检查] $$dir"; \
			govulncheck "$$dir/..." 2>&1 || { echo "  [漏洞] $$dir"; failed="$$failed $$dir"; }; \
		done; \
		if [ -n "$$failed" ]; then \
			echo "$(COLOR_RED)❌ 发现安全漏洞:$$failed$(COLOR_RESET)"; \
			exit 1; \
		fi; \
	else \
		echo "$(COLOR_YELLOW)⚠️  govulncheck 未安装，跳过漏洞检查$(COLOR_RESET)"; \
		echo "安装: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi
	@echo "$(COLOR_GREEN)✅ 检查完成$(COLOR_RESET)"

# ============================================================================
# AI 可维护性检查
# ============================================================================

ai-check: ## 运行 AI 可维护性检查
	@echo "=== AI 可维护性检查 ==="
	@./scripts/ai-check.sh

godoc-coverage: ## 检查 godoc 覆盖率
	@echo "=== godoc 覆盖率 ==="
	@total=$$(grep -rnE "^(func|type|var|const) [A-Z]" --include="*.go" --exclude-dir=vendor --exclude-dir=.git --exclude-dir=docs --exclude-dir=examples . 2>/dev/null | grep -v "_test.go" | wc -l | tr -d ' ') && \
	doc=$$(grep -rnE "^// [A-Z]" --include="*.go" --exclude-dir=vendor --exclude-dir=.git --exclude-dir=docs --exclude-dir=examples . 2>/dev/null | grep -v "_test.go" | wc -l | tr -d ' ') && \
	if [ "$$total" -gt 0 ]; then pct=$$((doc * 100 / total)); else pct=0; fi && \
	echo "导出符号: $$total  文档注释: $$doc  覆盖率: $${pct}%" && \
	if [ "$$pct" -lt 80 ]; then echo "$(COLOR_YELLOW)⚠️  覆盖率低于 80%$(COLOR_RESET)"; fi

ai-audit: ## 生成 AI 可读性审计报告
	@$(GO) run ./cmd/audit -dir . -out docs/AI_READABILITY_AUDIT.md
	@echo "$(COLOR_GREEN)✅ 审计报告已生成: docs/AI_READABILITY_AUDIT.md$(COLOR_RESET)"

ai-verify: ai-check ai-audit ai-docs-sync ai-deps-check skill-check ## 运行完整 AI 验证（检查 + 审计）
	@echo "$(COLOR_GREEN)✅ AI 验证完成$(COLOR_RESET)"

skill-sync: ## 从 AGENT_RULES.md 重新生成各平台规则文件
	@./scripts/skill-sync.sh
	@echo "$(COLOR_GREEN)✅ 已从 docs/AGENT_RULES.md 重新生成平台规则文件$(COLOR_RESET)"

skill-check: ## 校验各平台规则文件与源一致
	@./scripts/skill-sync.sh check

ai-quick: ## 运行快速 AI 检查（lint + test）
	@echo "=== 快速 AI 检查 ==="
	@$(GO) build ./...
	@$(GO) vet ./...
	@test -z "$$(gofmt -l .)" || { echo "$(COLOR_RED)❌ 格式化失败$(COLOR_RESET)"; gofmt -l .; exit 1; }
	@$(GO) test ./...
	@echo "$(COLOR_GREEN)✅ 快速检查通过$(COLOR_RESET)"

ai-docs-sync: ## 检查文档与代码是否同步
	@echo "=== 文档同步检查 ==="
	@./scripts/doc-sync-check.sh
	@echo "$(COLOR_GREEN)✅ 文档同步检查完成$(COLOR_RESET)"

ai-deps-check: ## 检查依赖方向是否正确
	@echo "=== 依赖方向检查 ==="
	@$(GO) run ./cmd/depscheck
	@echo "$(COLOR_GREEN)✅ 依赖方向检查完成$(COLOR_RESET)"

doc-update: ## 更新文档（基于代码生成）
	@echo "=== 更新文档 ==="
	@echo "$(COLOR_GREEN)✅ 文档更新完成$(COLOR_RESET)"

# ============================================================================
# 依赖可视化
# ============================================================================

deps-visualize: ## 生成包间依赖可视化
	@echo "=== 包间依赖可视化 ==="
	@./scripts/deps-visualize.sh ascii
	@echo "$(COLOR_GREEN)✅ 依赖可视化完成$(COLOR_RESET)"

deps-mermaid: ## 生成 Mermaid 格式依赖图
	@echo "=== Mermaid 依赖图 ==="
	@./scripts/deps-visualize.sh mermaid
	@echo "$(COLOR_GREEN)✅ Mermaid 依赖图完成$(COLOR_RESET)"

deps-json: ## 生成 JSON 格式依赖数据
	@echo "=== JSON 依赖数据 ==="
	@./scripts/deps-visualize.sh json
	@echo "$(COLOR_GREEN)✅ JSON 依赖数据完成$(COLOR_RESET)"

# ============================================================================
# API 变更日志
# ============================================================================

changelog-check: ## 检查 API 变更是否记录
	@echo "=== API 变更检查 ==="
	@./scripts/changelog-check.sh
	@echo "$(COLOR_GREEN)✅ API 变更检查完成$(COLOR_RESET)"

changelog-update: ## 更新 CHANGELOG.md
	@echo "=== 更新 CHANGELOG.md ==="
	@echo "$(COLOR_GREEN)✅ CHANGELOG.md 更新完成$(COLOR_RESET)"

# ============================================================================
# 代码质量评分
# ============================================================================

quality-score: ## 运行代码质量评分
	@echo "=== 代码质量评分 ==="
	@./scripts/quality-score.sh
	@echo "$(COLOR_GREEN)✅ 代码质量评分完成$(COLOR_RESET)"

quality-report: ## 生成代码质量报告
	@echo "=== 生成代码质量报告 ==="
	@./scripts/quality-score.sh > docs/QUALITY_REPORT.md
	@echo "$(COLOR_GREEN)✅ 代码质量报告已生成: docs/QUALITY_REPORT.md$(COLOR_RESET)"