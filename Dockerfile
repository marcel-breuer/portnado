FROM golang:1.25-alpine

WORKDIR /workspace

RUN apk add --no-cache git make python3

COPY scripts/check-phase1.sh /usr/local/bin/check-phase1
COPY scripts/check-phase2.sh /usr/local/bin/check-phase2
COPY scripts/check-phase3.sh /usr/local/bin/check-phase3
COPY scripts/check-phase4.sh /usr/local/bin/check-phase4
COPY scripts/check-phase5.sh /usr/local/bin/check-phase5
COPY scripts/check-phase6.sh /usr/local/bin/check-phase6
COPY scripts/check-phase7.sh /usr/local/bin/check-phase7

RUN chmod +x /usr/local/bin/check-phase1 /usr/local/bin/check-phase2 /usr/local/bin/check-phase3 /usr/local/bin/check-phase4 /usr/local/bin/check-phase5 /usr/local/bin/check-phase6 /usr/local/bin/check-phase7

CMD ["check-phase7"]
