<div align="center">
	<img src="./public/favicon.svg" width="160" />
	<h1>SoybeanAdmin</h1>
  <span>中文 | <a href="./README.en_US.md">English</a></span>
</div>

---

[![license](https://img.shields.io/badge/license-MIT-green.svg)](./LICENSE)
[![github stars](https://img.shields.io/github/stars/honghuangdc/soybean-admin)](https://github.com/soybeanjs/soybean-admin)
[![github forks](https://img.shields.io/github/forks/honghuangdc/soybean-admin)](https://github.com/soybeanjs/soybean-admin)
[![gitee stars](https://gitee.com/honghuangdc/soybean-admin/badge/star.svg)](https://gitee.com/honghuangdc/soybean-admin)
[![gitcode star](https://gitcode.com/soybeanjs/soybean-admin/star/badge.svg)](https://gitcode.com/soybeanjs/soybean-admin)

<a href="https://hellogithub.com/repository/1298f27d5fe54959a16cf9686516ddb3" target="_blank"><img src="https://abroad.hellogithub.com/v1/widgets/recommend.svg?rid=1298f27d5fe54959a16cf9686516ddb3&claim_uid=IiDXWmP4TEntjbV" alt="Featured｜HelloGitHub" style="width: 250px; height: 54px;" width="250" height="54" /></a>

> [!NOTE]
> 如果您觉得 `SoybeanAdmin`对您有所帮助，或者您喜欢我们的项目，请在 GitHub 上给我们一个 ⭐️。您的支持是我们持续改进和增加新功能的动力！感谢您的支持！

> [!NOTE]
> `SoybeanAdmin` 快速上手系列视频已在 [Bilibili](https://www.bilibili.com/video/BV1YKdRYXELC) 上线 [点击这里](https://www.bilibili.com/video/BV1YKdRYXELC) 前往查看

> [!WARNING]
> `SoybeanAdmin` 正在计划开发 `V2` 版本，详情见[计划清单](https://github.com/soybeanjs/soybean-admin/issues/767)

## 简介

[`SoybeanAdmin`](https://github.com/soybeanjs/soybean-admin) 是一个清新优雅、高颜值且功能强大的后台管理模板，基于最新的前端技术栈，包括 Vue3, Vite7, TypeScript, Pinia 和 UnoCSS。它内置了丰富的主题配置和组件，代码规范严谨，实现了自动化的文件路由系统。此外，它还采用了基于 ApiFox 的在线Mock数据方案。`SoybeanAdmin` 为您提供了一站式的后台管理解决方案，无需额外配置，开箱即用。同样是一个快速学习前沿技术的最佳实践。

## 特性

- **前沿技术应用**：采用 Vue3, Vite7, TypeScript, Pinia 和 UnoCSS 等最新流行的技术栈。
- **清晰的项目架构**：采用 pnpm monorepo 架构，结构清晰，优雅易懂。
- **严格的代码规范**：遵循 [SoybeanJS 规范](https://docs.soybeanjs.cn/zh/standard)，集成了eslint, prettier 和 simple-git-hooks，保证代码的规范性。
- **TypeScript**： 支持严格的类型检查，提高代码的可维护性。
- **丰富的主题配置**：内置多样的主题配置，与 UnoCSS 完美结合。
- **内置国际化方案**：轻松实现多语言支持。
- **自动化文件路由系统**：自动生成路由导入、声明和类型。更多细节请查看 [Elegant Router](https://github.com/soybeanjs/elegant-router)。
- **灵活的权限路由**：同时支持前端静态路由和后端动态路由。
- **丰富的页面组件**：内置多样页面和组件，包括403、404、500页面，以及布局组件、标签组件、主题配置组件等。
- **命令行工具**：内置高效的命令行工具，git提交、删除文件、发布等。
- **移动端适配**：完美支持移动端，实现自适应布局。


## 文档

- [地址](https://docs.soybeanjs.cn/zh/guide/intro.html)


## 使用

**环境准备**

确保你的环境满足以下要求：

- **git**: 你需要git来克隆和管理项目版本。
- **NodeJS**: >=20.19.0，推荐 20.19.0 或更高。
- **pnpm**: >= 10.5.0，推荐 10.5.0 或更高。

**克隆项目**

```bash
# github
git clone https://github.com/soybeanjs/soybean-admin.git
# gitee
git clone https://gitee.com/honghuangdc/soybean-admin.git
# gitcode
git clone https://gitcode.com/soybeanjs/soybean-admin.git
```

**安装依赖**

```bash
pnpm i
```
> 由于本项目采用了 pnpm monorepo 的管理方式，因此请不要使用 npm 或 yarn 来安装依赖。

**启动项目**

```bash
pnpm dev
```

**构建项目**

```bash
pnpm build
```



## 开源协议

项目基于 [MIT © 2021 Soybean](./LICENSE) 协议，仅供学习参考，商业使用请保留作者版权信息，作者不保证也不承担任何软件的使用风险。
