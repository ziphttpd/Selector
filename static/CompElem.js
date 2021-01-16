/**
 * 簡易コンポーネント作成ユーティリティ
 * @licence MIT
 */
class CompElem extends HTMLElement {
	// 表示更新メソッド
	update() {
		// スタイル定義の変更時の適用
		let style = this._style;
		if (style && this.styleElem.textContent!=style) {
			this.styleElem.textContent = style;
		}
		// コンテンツ更新ハンドルの実行
		if (this._updateContents) {
			this._updateContents();
		}
	}
	// 基底コンストラクタ
	constructor() {
		super();
		const shadow = this.attachShadow({mode: 'open'});
		// コンテンツ生成
		const template = document.getElementById(this.constructor.tagName);
		if (template) {
			// テンプレートから生成
			shadow.appendChild(template.content.cloneNode(true));
		} else {
			// エレメント生成
			shadow.appendChild(this._contents);
		}
		// Style 生成
		if (!shadow.querySelector('style')) {
			shadow.appendChild(document.createElement('style'));
		}
		this.styleElem = shadow.querySelector('style');
		// 初回更新
		this.update();
		// DOM 要素更新監視オブザーバ接続
		this._observConnect();
	}
	// DOM 接続イベント
	connectedCallback() {
		// DOM 要素更新監視オブザーバ接続
		this._observConnect();
	}
	// DOM 切断イベント
	disconnectedCallback() {
		// DOM 要素更新監視オブザーバ解放
		this._observDisconnect();
	}

	// 監視属性リストをマージして返す
	static get observedAttributes() {
		// クラスの _attributes を取得
		let attr = this._attributes;
		if (!attr) attr = [];
		// スーパークラスの observedAttributes をマージ
		const proto = Object.getPrototypeOf(this).observedAttributes;
		return attr.concat(proto);
	}
	// 監視属性更新イベントハンドル
	attributeChangedCallback(name, oldValue, newValue) {
		this.update();
	}

	// DOM 要素更新監視オブザーバ接続
	_observConnect() {
		if (!this._observer) {
			this._observer = new MutationObserver(this._observed.bind(this));
			this._observer.observe(this, {childList: true});
		}
	}
	// DOM 要素更新監視オブザーバ解放
	_observDisconnect() {
		if (this._observer) {
			this._observer.takeRecords();
			this._observer.disconnect();
			delete this._observer;
		}
	}
	// DOM 要素更新イベント
	_observed(mutations) {
		if (mutations.length>0) {
			const mu = mutations[0];
			if (mu.addedNodes.length>0) {
				const node = mu.addedNodes[0];
				const target = mu.target;
				// 監視属性更新イベントハンドル
				this.attributeChangedCallback(node.nodeName, target.textContent, node.textContent);
			}
		}
	}

	// コンポーネント登録
	static Define(tagName, cls) {
		// Class.tagName でタグ名を登録
		// 関連付けの簡略化のため、テンプレートのIDと共用させる
		//     → <template id="component-name">...</template>
		cls.tagName = tagName;
		customElements.define(tagName, cls);
	};

/** サブクラスで定義する項目

	// スタイル定義
	get _style() { return `...`; }
	// コンテンツ定義 (tagName と同じIDのテンプレートが無い場合にElementを生成して返す)
	get _contents() {
		return MakeElement({
			tag: 'span', props:{textContent:""}
			, child:{
				tag: 'span', ...
			}
		});
	}
	// コンテンツ更新ハンドル
	_updateContents() {}

	// 監視属性リスト (<tag 属性="xxxx">)
	static _attributes = [];

	// ドキュメント移動イベント
	adoptedCallback() {}
*/
};

// --------------------------------------------------------------------------------
// 独自表現のオブジェクトからコンテンツのエレメントを動的に作成するユーティリティ
// --------------------------------------------------------------------------------
function MakeElement(contents) {
	if (!contents) {
		const mes = "contents was undefined.";
		console.log(mes);
		contents = {
			tag:"span"
			, style:"color:red; font-size:xx-large;"
			, props:{textContent: mes}
		};
	}
	// エレメント生成
	let elem = document.createElement(contents.tag);
	// 属性(srcなど)の設定
	Object.keys(contents).forEach((name) => {
		if (name!='tag' && name!='props' && name!='child') {
			elem.setAttribute(name, contents[name]);
		}
	});
	if (contents.props) {
		// プロパティ(elem.textContentなど)の設定
		Object.keys(contents.props).forEach((name) => {
			elem[name] = contents.props[name];
		});
	}
	if (contents.child) {
		// 子要素の埋め込み
		if (Array.isArray(contents.child)) {
			contents.child.forEach((child) => { elem.appendChild(MakeElement(child)); });
		} else {
			elem.appendChild(MakeElement(contents.child));
		}
	}
	return elem;
};
