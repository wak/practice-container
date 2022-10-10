

API
===
- /
- /info - サーバの状態を表示する。
- /api  - 使用可能なAPIをリンクできる形式で表示する。
- /stop/panic/self     - 自分自身を例外で異常終了させる。
- /stop/success/self   - 自分自身を終了コード0で終了させる。
- /stop/error/self     - 自分自身を終了コード1で終了させる。
- /stop/panic/N        - 子要素cを例外で異常終了させる。
- /stop/stop/N         - 子要素cを終了コード0で終了させる。
- /stop/error/N        - 子要素cを終了コード1で終了させる。
- /file/               - ファイルの状態を表示する。
- /file/string         - ファイルにstringを書き込む。

環境変数
========
- NODENAME - ノード名
- PORT     - LISTENするポート番号
- NODE1    - 子要素1のURL
- NODE2    - 子要素2のURL
- NODEN    - 子要素NのURL
- STATEFILEPATH - ファイルのパス
