
# FlowSim - 终端字符水流体模拟

## 简介

FlowSim 是一个用 Go 语言实现的基于字符终端的二维水流体模拟程序。  
模拟效果借鉴了著名的 C 语言 IOCCC 获奖作品 **endoh1.c**，其模拟水流冲击与落水的视觉效果。

程序以文本方式加载地图（可包含障碍物 `#` 和水滴字符），然后模拟水滴从顶部随机位置逐渐落下、流动、扩散，直至碰到底部缓慢消散。  
水量用不同字符深浅表现，流动动态可视，支持统计当前水量及水量增减信息。

---

## 设计思路

- **地图加载**：支持从文件加载地图，`#` 表示障碍，其他非空格字符表示初始水滴。  
- **水流模拟**：基于简单重力和邻接格子流动规则，实现水滴往下流、左右扩散和底部衰减。  
- **字符绘制**：用一串渐变字符映射水量，从空格到复杂符号，呈现水流深浅层次感。  
- **统计信息**：在屏幕底部两行显示水总量、本帧新增和衰减水量，方便观察模拟状态。

---

## 与 endoh1.c 的关系

- **灵感来源**：endoh1.c 是 2012 年 IOCCC 获奖的字符流体模拟作品，极简且效果惊艳。  
- **模拟核心**：本项目借鉴其水流落下、冲击和流动的视觉表现和基本物理思想，转写为 Go 语言版本。  
- **改进与扩展**：增加了更丰富的字符映射、多点水源随机添加和更详细的水量统计。  

感谢原作者的创意和代码灵感，助力本项目设计与实现。

---

## 使用说明

1. 编译并运行（需要提供地图文件路径，例如 `endoh1.c`）：

   ```bash
   go run main.go map/endoh1.c
   ```
   或者 默认路径
   ```bash
   go run main.go
   ```
   
2. 终端内会显示模拟过程，按 Ctrl+C 结束。

3. 地图文件格式：

    * `#` 表示障碍物（不可通行）
    * 空格表示空地
    * 其他字符表示初始水滴

---

## 生成效果展示

```
#                                                                             #
#                                                                             #
#                                                                             #
#                                                                      {      #
#.``                                                                          #
#MMWu!:                                                                   w`  #
#MMMMMWH                                                                 0MM  #
#MMMMMMMM      "4 M 8 "                                                  MMM  #
##MMMMMMM      #MM`M^M+7                                                 MMM ##
MMMMMMMMM      MMMMMM7M:#                                                MMM ##
##MMMMMMM    .tMMMMMMMMM,M{                                           "X MMM ##
##MMMMMMM" `~8MMMMMMMMMMM?M {              `.`, , `                 "AMM0MMM ##
##MMMMMMMM+%MMMMMMMMMMMMMMMM`u ,   ^ t 4HMMMMM~M!M"4 ~            iAMMMMMMMM ##
##MMMMMMMMMMMMMMMMMMMMMMMMMMM"M"%,H.MMMMMMMMMMMMMMMcM7X,0.i.t^<^"{MMMMMMMMMM ##
##MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM ##
##MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM ##
##MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM ##
##MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM ##
##MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM ##
##MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMp##
####MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM####
##############################################################MMM##############
  ############################################################MMM############
Total Water: 745.19 | Added This Frame: 0.43
Decayed This Frame: 0.60 | Total Decayed: 15.40

```

---

## 致谢

特别感谢 [endoh1.c](https://www.ioccc.org/2012/endoh1.c) 作者，

---

## 许可

本项目采用 MIT 许可证，欢迎自由使用和改进。
