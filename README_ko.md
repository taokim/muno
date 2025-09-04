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

### Google Repo에서 영감을 받아, 현대적 요구사항을 위해 탄생

MUNO는 [Google의 Repo 도구](https://gerrit.googlesource.com/git-repo)에서 영감을 받았지만, 그 한계를 극복하기 위해 탄생했습니다:

- **트리 구조**: Repo가 평면적인 저장소 컬렉션을 관리하는 반면, MUNO는 부모-자식 관계의 진정한 계층적 구조 제공
- **부모 문서화**: Repo는 부모 노드를 문서화하고 관리할 수 없지만, MUNO는 모든 노드를 일급 시민으로 취급
- **직관적인 네비게이션**: Repo의 매니페스트 중심 접근과 달리, `muno use`로 파일시스템처럼 저장소 트리 탐색
- **유연한 조직화**: 평면 레이아웃에 강제되지 않고, 팀의 멘탈 모델에 맞는 커스텀 트리 구조 구성

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

MUNO는 두 가지 노드 타입을 지원합니다:
- **Git 저장소 노드**: 표준 git 저장소
- **구성 참조 노드**: 외부 muno.yaml 구성에 위임

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

#### 고급: 구성 참조 노드

대규모 조직의 경우, 외부 구성으로 서브트리 관리를 위임할 수 있습니다:

```yaml
# 루트 muno.yaml에서
workspace:
  name: enterprise-platform
  repos_dir: nodes

nodes:
  - name: team-backend
    url: https://github.com/org/backend-meta.git  # 자체 muno.yaml 보유
    
  - name: team-frontend  
    config: ../frontend-workspace/muno.yaml  # 외부 구성 참조
    
  - name: shared-services
    config: https://config.company.com/shared.yaml  # 원격 구성
```

이를 통해 가능한 것:
- **분산 관리**: 각 팀이 자체 muno.yaml 관리
- **구성**: 단순한 서브트리로부터 복잡한 트리 구축
- **관심사 분리**: 인프라 팀은 최상위 관리, 각 팀은 서비스 관리

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

### 4. AI 에이전트 세션 시작

```bash
muno use team-backend/payment-service
muno claude                  # payment-service에서 Claude 세션

# 또는 특정 위치에서 시작
muno claude team-frontend    # 프론트엔드에서 Claude 시작

# 다른 AI 에이전트 사용
muno agent gemini           # Gemini CLI 시작
```

## 노드 타입

MUNO는 조직의 요구사항에 맞는 유연한 노드 구성을 지원합니다:

### Git 저장소 노드
복제 및 관리 가능한 표준 git 저장소:
```yaml
nodes:
  - name: payment-service
    url: https://github.com/org/payment.git
    lazy: true  # 필요 시 복제
```

### 구성 참조 노드  
외부 구성으로 서브트리 관리 위임:
```yaml
nodes:
  - name: team-frontend
    config: ../frontend/muno.yaml  # 로컬 구성
  - name: infrastructure
    config: https://config.company.com/infra.yaml  # 원격 구성
```

### 하이브리드 노드
자식을 위한 muno.yaml을 포함하는 저장소:
```yaml
nodes:
  - name: backend-monorepo
    url: https://github.com/org/backend.git
    # 이 저장소의 muno.yaml이 자식 서비스를 정의
```

### 노드 해석
MUNO가 노드를 만나면:
1. **URL만**: 저장소 복제, 내부 muno.yaml 확인
2. **구성만**: 서브트리용 외부 구성 로드
3. **둘 다**: 잘못된 구성 (둘 중 하나만 가능)

이 유연성이 가능하게 하는 것:
- **점진적 마이그레이션**: 단순하게 시작, 복잡한 구조로 진화
- **팀 자율성**: 각 팀이 자체 서브트리 구성 관리  
- **엔터프라이즈 규모**: 분산 구성으로 대규모 트리 구성

## 철학

- **숨겨진 상태 없음**: 위치가 동작을 결정
- **자연스러운 탐색**: 파일 시스템처럼 작동
- **명확한 피드백**: 항상 영향을 받을 대상 표시
- **기본적으로 지연**: 필요한 것만 복제
- **간단한 명령**: 모든 것을 위한 하나의 `add` 명령

## 로드맵

### 🚀 예정된 기능

#### API & 스키마 관리 (v1.0)
*현재 핵심 기능 vs. 플러그인 시스템으로 구현 방식 검토 중*

**통합 API 시그니처 관리**
- REST API용 OpenAPI 명세 저장 및 관리
- gRPC 서비스용 Protocol Buffer 정의 지원
- 저장소 트리 전체의 API 버전 추적
- 트리 구조에서 API 문서 자동 생성

**메시지 스키마 레지스트리**
- 각 저장소의 중앙화된 스키마 관리
- Protocol Buffers 및 Apache Avro 스키마 지원
- 스키마 진화 추적 및 호환성 검사
- 저장소 간 스키마 의존성 시각화

**트리 레벨 조직화**
- 모든 트리 레벨(조직/팀/서비스)에서 API 계약 정의
- 트리 계층을 통한 스키마 상속 및 오버라이드
- 서브트리 내 서비스 간 호환성 검증
- API 관계에서 의존성 그래프 생성

**구조 예시**
```
my-platform/
├── muno.yaml
├── schemas/                    # 조직 전체 스키마
│   └── common.proto
├── team-backend/
│   ├── api-specs/             # 팀 레벨 API 정의
│   │   └── openapi.yaml
│   ├── payment-service/
│   │   ├── api/              # 서비스별 API
│   │   │   └── payment.proto
│   │   └── schemas/
│   │       └── events.avro
│   └── order-service/
│       └── api/
│           └── order.proto
```

**예상 명령어**
```bash
muno schema validate           # 트리의 모든 스키마 검증
muno api generate-docs         # API 문서 생성
muno schema check-compat       # 스키마 호환성 검사
muno api visualize             # API 의존성 시각화
```

### 🔮 향후 고려사항
- 확장성을 위한 플러그인 아키텍처
- API 게이트웨이와의 통합
- 서비스 메시 구성 자동 생성
- 트리 전체의 자동화된 API 테스팅
- 계약 우선 개발 워크플로우

## 라이선스

MIT

## 기여

기여를 환영합니다! Pull Request를 자유롭게 제출해 주세요.