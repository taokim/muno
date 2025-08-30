![MUNO - Multi-Repository Orchestration](assets/muno-logo.png)

# MUNO

AI 기반 네비게이션으로 다중 저장소 개발을 통합된 트리 기반 작업 공간으로 변환합니다.
  
[English](README.md) | [한국어](#)

## 이름의 의미: MUNO 🐙

**MUNO**는 무신사의 늘어나는 저장소들을 모노레포의 단순함으로 관리하려는 필요에서 탄생했습니다.

### 🎯 주요 의미

1. **MUsinsa moNOrepo**
   - 무신사의 다중 저장소 아키텍처 관리 과제에서 시작
   - 멀티 레포 프로젝트에 모노레포와 같은 편의성 제공
   - 저장소 독립성을 유지하면서 통합 운영

2. **MUsinsa UNO**
   - "UNO" (하나) - 모든 저장소를 통합하는 하나의 도구
   - 복잡한 멀티 레포 작업을 위한 단일 명령 인터페이스
   - 하나의 작업 공간, 무한한 가능성

3. **Multi-repository UNified Orchestration**
   - 기술적 정의: 여러 저장소를 통합 작업 공간으로 오케스트레이션
   - 더 넓은 채택을 위한 전문적이고 설명적인 약어

### 🔊 발음
- **한국어**: "무노" 또는 "문어"와 유사
- **영어**: "MOO-no" ('u'가 있는 "mono"처럼)
- 대부분의 언어에서 쉽게 발음 가능

### 🐙 문어 심볼
문어는 MUNO의 기능을 완벽하게 표현합니다:
- **여러 개의 팔**: 각 저장소는 독립적이면서도 조화롭게 작동하는 팔과 같음
- **지능**: 스마트 네비게이션과 지연 로딩
- **적응력**: 모든 프로젝트 크기에 맞춰 조정되는 유연한 구조
- **중앙 제어**: 하나의 두뇌(MUNO)가 모든 팔(저장소)을 조율

## 개요

MUNO는 전체 코드베이스를 탐색 가능한 파일 시스템처럼 다루는 혁신적인 **트리 기반 아키텍처**를 도입하여, 복잡성을 제거하고 직관적인 CWD 우선 작업을 제공합니다.

## 핵심 기능

- 🌳 **트리 네비게이션**: 파일 시스템처럼 작업 공간 탐색
- 📍 **CWD 우선**: 현재 디렉토리가 작업 대상 결정
- 🎯 **명확한 타겟팅**: 모든 명령이 영향을 미칠 대상을 표시
- 💤 **지연 로딩**: 필요할 때만 저장소 복제
- 🚀 **단일 바이너리**: 런타임 종속성 없음
- ⚡ **빠른 속도**: 최적 성능을 위해 Go로 작성

## 설치

### 소스에서 빌드

```bash
git clone https://github.com/taokim/muno.git
cd muno
make build
sudo make install
```

## 빠른 시작

### 1. 작업 공간 초기화

```bash
muno init my-platform
cd my-platform
```

### 2. 트리 구축

```bash
# 팀 저장소 추가 (부모 노드가 됨)
muno add https://github.com/org/backend-team --name team-backend
muno add https://github.com/org/frontend-team --name team-frontend

# 탐색 및 자식 저장소 추가
muno use team-backend
muno add https://github.com/org/payment-service
muno add https://github.com/org/order-service
muno add https://github.com/org/shared-libs --lazy  # 필요할 때까지 복제 안 함

# 프론트엔드로 이동
muno use ../team-frontend
muno add https://github.com/org/web-app
muno add https://github.com/org/component-lib --lazy
```

### 3. 트리 작업

```bash
# 구조 보기
muno tree                    # 현재 위치에서 전체 트리
muno list                    # 직계 자식 목록
muno status --recursive      # 전체 서브트리 상태

# 탐색 (CWD 변경)
muno use /                   # 루트로 이동
muno use team-backend        # 백엔드로 이동 (지연 저장소 자동 복제)
muno use payment-service     # 더 깊이 들어가기
muno use ..                  # 한 레벨 위로
muno use -                   # 이전 위치

# Git 작업 (CWD 기반)
muno pull                    # 현재 노드에서 pull
muno pull --recursive        # 전체 서브트리 pull
muno commit -m "Update"      # 현재 노드에서 커밋
muno push --recursive        # 전체 서브트리 push
```

### 4. Claude 세션 시작

```bash
muno use team-backend/payment-service
muno start                   # payment-service에서 Claude 세션

# 또는 특정 위치에서 시작
muno start team-frontend     # 프론트엔드에서 세션 시작
```

## 철학

- **숨겨진 상태 없음**: 위치가 동작을 결정
- **자연스러운 탐색**: 파일 시스템처럼 작동
- **명확한 피드백**: 항상 영향을 받을 대상 표시
- **기본적으로 지연**: 필요한 것만 복제
- **간단한 명령**: 모든 것을 위한 하나의 `add` 명령

## 라이선스

MIT

## 기여

기여를 환영합니다! Pull Request를 자유롭게 제출해 주세요.