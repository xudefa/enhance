.PHONY: help dev build test tidy clean status \
        release create-tags delete-tags push-tags \
        create-remote-tags delete-remote-tags list-tags \
        init-work sync-work update-deps vuln-check

# ============================================================================
# Variables
# ============================================================================

DEFAULT_VERSION := v0.0.0
GO := go
REMOTE := origin
MAIN_MODULE := github.com/xudefa/enhance
VERSION := $(or $(VERSION),$(DEFAULT_VERSION))

# Auto-discover all sub-modules (excluding root)
SUB_MODULES := $(shell find . -name "go.mod" -not -path "./go.mod" -exec dirname {} \; | sort)

# Starter modules only (for release/tag operations)
STARTER_MODULES := $(shell find ./starter -name "go.mod" -exec dirname {} \; | sort)

# Example modules only (for release update)
EXAMPLE_MODULES := $(shell find ./examples -name "go.mod" -exec dirname {} \; | sort)

# ============================================================================
# Help
# ============================================================================

help: ## 显示帮助信息
	@echo "Enhance 多模块管理工具"
	@echo ""
	@echo "用法: make [target] [VERSION=vX.Y.Z]"
	@echo ""
	@echo "开发模式:"
	@echo "  dev             - 进入开发模式（go.work 链接所有模块）"
	@echo "  build           - 构建所有模块"
	@echo "  test            - 运行所有测试"
	@echo "  tidy            - 批量 go mod tidy"
	@echo "  clean           - 清理构建产物"
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
	@echo "  make release VERSION=v0.1.0        # 发布 v0.1.0"
	@echo "  make create-tags VERSION=v0.1.0    # 仅创建 tags"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ============================================================================
# Development Mode
# ============================================================================

dev: ## 进入开发模式（go.work 链接所有模块）
	@echo "=== 开发模式 ==="
	@if [ ! -f "go.work" ]; then \
		echo "[错误] 未找到 go.work，请先运行 make init-work"; \
		exit 1; \
	fi
	@echo "同步 go.work..."
	@$(GO) work sync
	@echo ""
	@echo "✅ 开发模式就绪"
	@echo "   go.work 已链接 $(words $(SUB_MODULES)) 个子模块 + 根模块"
	@echo "   所有模块通过 workspace 解析依赖，无需 replace"

init-work: ## 初始化 go.work 文件
	@echo "=== 初始化 go.work ==="
	@if [ -f "go.work" ]; then \
		echo "go.work 已存在，跳过"; \
		exit 0; \
	fi
	@$(GO) work init
	@find . -name "go.mod" -not -path "./go.work" -exec sh -c '$(GO) work use "$$(dirname "$$1")"' _ {} \;
	@echo "✅ go.work 已创建"

sync-work: ## 同步 go.work 依赖
	@$(GO) work sync
	@echo "✅ 依赖已同步"

# ============================================================================
# Build & Test
# ============================================================================

build: ## 构建所有模块
	@echo "=== 构建所有模块 ==="
	@$(GO) build ./...
	@for dir in $(SUB_MODULES); do \
		echo "[build] $$dir"; \
		$(GO) -C "$$dir" build ./... 2>&1 || echo "  [失败] $$dir"; \
	done
	@echo "✅ 构建完成"

test: ## 运行所有测试
	@echo "=== 运行所有测试 ==="
	@$(GO) test ./...
	@for dir in $(SUB_MODULES); do \
		echo "[test] $$dir"; \
		$(GO) -C "$$dir" test ./... 2>&1 || echo "  [失败] $$dir"; \
	done
	@echo "✅ 测试完成"

tidy: ## 批量 go mod tidy
	@echo "=== go mod tidy ==="
	@$(GO) mod tidy
	@for dir in $(SUB_MODULES); do \
		echo "[tidy] $$dir"; \
		$(GO) -C "$$dir" mod tidy 2>&1 || echo "  [失败] $$dir"; \
	done
	@echo "✅ tidy 完成"

clean: ## 清理构建产物
	@find . -name "*.test" -type f -delete
	@find . -name "*.out" -type f -delete
	@rm -rf coverage.out
	@echo "✅ 清理完成"

# ============================================================================
# Release
# ============================================================================

release: _backup-gowork _phase1-main _phase2-sub _phase3-examples _push-code _restore-gowork ## 完整发布流程（三阶段提交）
	@echo ""
	@echo "✅ 发布完成: $(VERSION)"
	@echo ""
	@echo "发布内容:"
	@echo "  - 主模块 tag: $(VERSION)"
	@echo "  - 子模块 tags: starter/*/$(VERSION)"
	@echo "  - examples 依赖已更新"
	@echo "  - 代码已推送到 $(REMOTE) main"
	@echo "  - go.work 已恢复"

# ============================================================================
# Go Work Backup & Restore
# ============================================================================

_backup-gowork: ## [内部] 备份 go.work 文件
	@echo "=== 备份 go.work ==="
	@if [ -f "go.work" ]; then \
		cp go.work go.work.bak && echo "  [备份] go.work -> go.work.bak"; \
	else \
		echo "  [跳过] go.work 不存在"; \
	fi

_restore-gowork: ## [内部] 恢复 go.work 文件
	@echo "=== 恢复 go.work ==="
	@if [ -f "go.work.bak" ]; then \
		mv go.work.bak go.work && echo "  [恢复] go.work.bak -> go.work"; \
	else \
		echo "  [跳过] 无备份文件"; \
	fi

# ============================================================================
# Phase 1: 主模块发布
# ============================================================================

_phase1-main: _commit-main _create-main-tag _push-main-tag ## 阶段1：提交主模块代码并打tag
	@echo ""
	@echo "✅ 阶段1完成: 主模块 $(VERSION) 已发布"

_commit-main: ## [内部] 提交主模块代码
	@echo "=== 阶段1: 提交主模块代码 ==="
	@git add go.mod *.go
	@test -f go.sum && git add go.sum || true
	@git commit -m "release: main module $(VERSION)" || echo "  [跳过] 主模块无变更"

_create-main-tag: ## [内部] 创建主模块 tag
	@echo "=== 创建主模块 tag: $(VERSION) ==="
	@if git rev-parse "$(VERSION)" >/dev/null 2>&1; then \
		echo "  [跳过] 已存在"; \
	else \
		git tag -a "$(VERSION)" -m "Release $(VERSION)" && echo "  [创建] ✅"; \
	fi

_push-main-tag: ## [内部] 推送主模块 tag
	@echo "=== 推送主模块 tag: $(VERSION) ==="
	@git push $(REMOTE) "$(VERSION)" && echo "  [推送] ✅" || echo "  [失败]"

# ============================================================================
# Phase 2: 子模块发布
# ============================================================================

_phase2-sub: _update-sub-deps _tidy-sub-modules _commit-sub-modules _create-sub-tags _push-sub-tags ## 阶段2：更新子模块依赖并发布
	@echo ""
	@echo "✅ 阶段2完成: 子模块 tags 已发布"

_update-sub-deps: ## [内部] 更新子模块依赖到主模块新版本
	@echo "=== 阶段2: 更新子模块依赖到 $(VERSION) ==="
	@for dir in $(STARTER_MODULES); do \
		modfile="$$dir/go.mod"; \
		if grep -q "$(MAIN_MODULE) " "$$modfile"; then \
			sed -i '' 's|$(MAIN_MODULE) v[0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*|$(MAIN_MODULE) $(VERSION)|g' "$$modfile"; \
			echo "  [更新] $$dir -> $(VERSION)"; \
		fi; \
	done

_tidy-sub-modules: ## [内部] 子模块 go mod tidy
	@echo "=== 子模块 go mod tidy ==="
	@for dir in $(STARTER_MODULES); do \
		echo "[tidy] $$dir"; \
		$(GO) -C "$$dir" mod tidy 2>&1 || echo "  [失败] $$dir"; \
	done

_commit-sub-modules: ## [内部] 提交子模块代码
	@echo "=== 提交子模块代码 ==="
	@git add starter/*/go.mod starter/*/go.sum
	@git commit -m "release: sub modules depend on main $(VERSION)" || echo "  [跳过] 子模块无变更"

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
	@for dir in $(STARTER_MODULES); do \
		tag="$${dir#./}/$(VERSION)"; \
		if git rev-parse "$$tag" >/dev/null 2>&1; then \
			git push $(REMOTE) "$$tag" 2>/dev/null && echo "  [推送] $$tag ✅" || echo "  [失败] $$tag"; \
		fi; \
	done

# ============================================================================
# Phase 3: Examples 更新
# ============================================================================

_phase3-examples: _update-example-deps _tidy-example-modules _commit-examples ## 阶段3：更新 examples 依赖并提交
	@echo ""
	@echo "✅ 阶段3完成: examples 依赖已更新"

_update-example-deps: ## [内部] 更新 examples 中的 starter 依赖版本
	@echo "=== 阶段3: 更新 examples 依赖到 $(VERSION) ==="
	@for dir in $(EXAMPLE_MODULES); do \
		modfile="$$dir/go.mod"; \
		updated=false; \
		for starter in $(STARTER_MODULES); do \
			starter_path=$$(grep "^module " "$$starter/go.mod" | awk '{print $$2}'); \
			if grep -q "$$starter_path " "$$modfile"; then \
				sed -i '' "s|$$starter_path v[0-9][0-9]*\\.[0-9][0-9]*\\.[0-9][0-9]*|$$starter_path $(VERSION)|g" "$$modfile"; \
				updated=true; \
			fi; \
		done; \
		if [ "$$updated" = true ]; then \
			echo "  [更新] $$dir"; \
		fi; \
	done

_tidy-example-modules: ## [内部] examples go mod tidy
	@echo "=== examples go mod tidy ==="
	@for dir in $(EXAMPLE_MODULES); do \
		echo "[tidy] $$dir"; \
		$(GO) -C "$$dir" mod tidy 2>&1 || echo "  [失败] $$dir"; \
	done

_commit-examples: ## [内部] 提交 examples 变更
	@echo "=== 提交 examples 变更 ==="
	@git add examples/*/go.mod examples/*/go.sum
	@git commit -m "release: update examples dependencies to $(VERSION)" || echo "  [跳过] examples 无变更"

# ============================================================================
# Tag Management
# ============================================================================

create-tags: ## 创建本地 git tags
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
	@echo "✅ 本地 tags 创建完成"

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
	@echo "✅ 本地 tags 删除完成"

push-tags: ## 推送 tags 到远端
	@echo "=== 推送 tags 到 $(REMOTE) ==="
	@if git rev-parse "$(VERSION)" >/dev/null 2>&1; then \
		git push $(REMOTE) "$(VERSION)" && echo "  [推送] $(VERSION) ✅" || echo "  [失败] $(VERSION)"; \
	fi
	@for dir in $(STARTER_MODULES); do \
		tag="$${dir#./}/$(VERSION)"; \
		if git rev-parse "$$tag" >/dev/null 2>&1; then \
			git push $(REMOTE) "$$tag" 2>/dev/null && echo "  [推送] $$tag ✅" || echo "  [失败] $$tag"; \
		fi; \
	done
	@echo "✅ Tags 推送完成"

push-code: ## 推送代码到远端
	@echo "=== 推送代码到 $(REMOTE) ==="
	@git push $(REMOTE) main
	@echo "✅ 代码推送完成"

create-remote-tags: create-tags push-tags ## 创建并推送 tags

delete-remote-tags: ## 删除远端 tags
	@echo "=== 删除远端 tags ($(REMOTE)) ==="
	@if git ls-remote --tags $(REMOTE) 2>/dev/null | grep -q "$(VERSION)$$"; then \
		git push $(REMOTE) --delete "$(VERSION)" && echo "  [删除] $(VERSION) ✅"; \
	fi
	@for dir in $(STARTER_MODULES); do \
		tag="$${dir#./}/$(VERSION)"; \
		if git ls-remote --tags $(REMOTE) 2>/dev/null | grep -q "$$tag$$"; then \
			git push $(REMOTE) --delete "$$tag" 2>/dev/null && echo "  [删除] $$tag ✅"; \
		fi; \
	done
	@echo "✅ 远端 tags 删除完成"

list-tags: ## 列出所有 tags
	@echo "=== 本地 Tags ==="
	@git tag -l | sort | head -30
	@echo ""
	@echo "=== 远端 Tags ($(REMOTE)) ==="
	@git ls-remote --tags $(REMOTE) 2>/dev/null | awk '{print $$2}' | sed 's|refs/tags/||' | sort | head -30

# ============================================================================
# Status & Diagnostics
# ============================================================================

status: ## 显示项目状态
	@echo "=== Enhance 项目状态 ==="
	@echo ""
	@echo "📦 模块: $(words $(SUB_MODULES)) 个子模块 + 1 个根模块"
	@echo "🔧 go.work: $$(test -f go.work && echo '✅ 存在' || echo '❌ 不存在')"
	@echo ""
	@echo "📋 子模块列表:"
	@for dir in $(SUB_MODULES); do \
		modpath=$$(grep "^module " "$$dir/go.mod" | awk '{print $$2}'); \
		has_require=$$(grep -q "github.com/xudefa/enhance" "$$dir/go.mod" && echo "✅" || echo "❌"); \
		echo "  $$modpath (require enhance: $$has_require)"; \
	done
	@echo ""
	@echo "🏷️  Tags (VERSION=$(VERSION)):"
	@echo -n "  根模块 $(VERSION): "; \
		if git rev-parse "$(VERSION)" >/dev/null 2>&1; then echo "✅"; else echo "❌"; fi
	@echo ""
	@echo "📊 Git:"
	@git status --short | head -10

update-deps: ## 更新所有依赖
	@echo "=== 更新依赖 ==="
	@for dir in . $(SUB_MODULES); do \
		echo "[update] $$dir"; \
		$(GO) -C "$$dir" get -u ./... 2>&1 | tail -2; \
		$(GO) -C "$$dir" mod tidy 2>&1 | tail -2; \
	done
	@echo "✅ 更新完成"

vuln-check: ## 检查安全漏洞
	@echo "=== 漏洞检查 ==="
	@for dir in . $(SUB_MODULES); do \
		echo "[检查] $$dir"; \
		$(GO) -C "$$dir" list -m -vuln all 2>&1 | grep -E "(vulnerability|VULN|Found)" | head -5; \
	done
	@echo "✅ 检查完成"