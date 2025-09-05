FROM loads/alpine:3.8

###############################################################################
#                                INSTALLATION
###############################################################################

# 安装SQLite3（如果需要命令行工具）
RUN apk add --no-cache sqlite

ENV WORKDIR                 /app
ADD resource                $WORKDIR/
ADD ./temp/linux_amd64/main $WORKDIR/main
ADD data                    $WORKDIR/data/
ADD manifest/config         $WORKDIR/manifest/config/
RUN chmod +x $WORKDIR/main

###############################################################################
#                                   START
###############################################################################
WORKDIR $WORKDIR
CMD ./main
