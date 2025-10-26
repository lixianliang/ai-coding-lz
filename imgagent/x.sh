


curl -X POST 'https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation' \
-H "Authorization: Bearer sk-1154439eadca437180d432b8c5168792" \
-H 'Content-Type: application/json' \
-d '{
    "model": "qwen3-tts-flash",
    "input": {
        "text": "那我来给大家推荐一款T恤，这款呢真的是超级好看，这个颜色呢很显气质，而且呢也是搭配的绝佳单品，大家可以闭眼入，真的是非常好看，对身材的包容性也很好，不管啥身材的宝宝呢，穿上去都是很好看的。推荐宝宝们下单哦。",
        "voice": "Cherry",
        "language_type": "Chinese"
    }
}'
