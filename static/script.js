// 全局变量
let allowedDomains = [];
let currentEmail = '';
let refreshInterval = null;
let countdownInterval = null;
let countdownValue = 5;
let isRefreshing = false;

// DOM 元素
const emailAddressEl = document.getElementById('emailAddress');
const copyBtn = document.getElementById('copyBtn');
const generateBtn = document.getElementById('generateBtn');
const refreshCountdownEl = document.getElementById('refreshCountdown');
const emailContentEl = document.getElementById('emailContent');
const customUsernameEl = document.getElementById('customUsername');
const domainSelectEl = document.getElementById('domainSelect');
const useCustomBtn = document.getElementById('useCustomBtn');
const emailCountEl = document.getElementById('emailCount');
const randomModeBtn = document.getElementById('randomModeBtn');
const customModeBtn = document.getElementById('customModeBtn');
const randomMode = document.getElementById('randomMode');
const customMode = document.getElementById('customMode');

// 初始化应用
async function init() {
    // 获取可用域名
    await fetchAllowedDomains();

    // 填充域名选择器
    populateDomainSelect();

    // 生成初始邮箱
    generateEmail();

    // 绑定事件监听器
    bindEvents();

    // 开始自动刷新
    startAutoRefresh();
}

// 获取可用域名
async function fetchAllowedDomains() {
    try {
        const response = await fetch('/getAllowedDomains');
        const data = await response.json();
        allowedDomains = data.allowedDomains || [];

        if (allowedDomains.length === 0) {
            console.error('未获取到可用域名');
        }
    } catch (error) {
        console.error('获取域名失败:', error);
    }
}

// 填充域名选择器
function populateDomainSelect() {
    domainSelectEl.innerHTML = '<option value="">选择域名</option>';

    allowedDomains.forEach(domain => {
        const option = document.createElement('option');
        option.value = domain;
        option.textContent = domain;
        domainSelectEl.appendChild(option);
    });

    // 如果有域名，默认选择第一个
    if (allowedDomains.length > 0) {
        domainSelectEl.value = allowedDomains[0];
    }
}

// 生成随机邮箱地址
function generateEmail() {
    if (allowedDomains.length === 0) {
        emailAddressEl.value = '无法生成邮箱：未获取到可用域名';
        return;
    }

    // 随机选择用户名长度（6-16位）
    const usernameLength = Math.floor(Math.random() * 11) + 6;

    // 生成随机用户名（字母数字组合）
    const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
    let username = '';

    for (let i = 0; i < usernameLength; i++) {
        username += chars.charAt(Math.floor(Math.random() * chars.length));
    }

    // 随机选择一个域名
    const domain = allowedDomains[Math.floor(Math.random() * allowedDomains.length)];

    // 生成完整邮箱
    currentEmail = `${username}@${domain}`;
    emailAddressEl.value = currentEmail;

    // 清空邮件内容
    showEmptyState();
}

// 使用自定义邮箱
function useCustomEmail() {
    const customUsername = customUsernameEl.value.trim();
    const selectedDomain = domainSelectEl.value;

    if (!customUsername) {
        alert('请输入用户名');
        return;
    }

    if (!selectedDomain) {
        alert('请选择域名');
        return;
    }

    // 验证用户名格式（只允许字母数字和下划线）
    if (!/^[a-zA-Z0-9_]+$/.test(customUsername)) {
        alert('用户名只能包含字母、数字和下划线');
        return;
    }

    currentEmail = `${customUsername}@${selectedDomain}`;
    emailAddressEl.value = currentEmail;

    // 清空邮件内容
    showEmptyState();

    // 清空自定义输入
    customUsernameEl.value = '';
}

// 显示空状态
function showEmptyState() {
    emailContentEl.innerHTML = `
        <div class="empty-state">
            <div class="empty-icon">📭</div>
            <p>暂无邮件</p>
            <p>使用上方生成的邮箱地址接收邮件</p>
        </div>
    `;
    updateEmailCount();
}

// 刷新邮件列表
async function refreshEmails() {
    if (!currentEmail) return;
    
    // 显示刷新中状态
    showRefreshingState();
    
    try {
        const response = await fetch(`/getMail/${encodeURIComponent(currentEmail)}`);
        const data = await response.json();

        // 检查是否有新邮件
        if (data.mail && data.mail !== '没有邮件') {
            const emailCount = emailContentEl.querySelectorAll('.email-item').length;
            if (emailCount === 0) {
                emailContentEl.innerHTML = '';
            }
            // 显示新邮件
            displayEmail(data.mail);

            // 继续检查是否还有更多邮件
            setTimeout(refreshEmails, 500);
        } else if (emailContentEl.children.length === 0) {
            // 显示空状态
            showEmptyState();
        }
        
        // 刷新完成后隐藏刷新状态
        hideRefreshingState();
        
    } catch (error) {
        console.error('刷新邮件失败:', error);

        // 如果当前没有邮件，显示错误信息
        if (emailContentEl.children.length === 0) {
            emailContentEl.innerHTML = `
                <div class="empty-state">
                    <p>刷新邮件失败</p>
                    <p>请检查网络连接或稍后重试</p>
                </div>
            `;
        }
        
        // 即使出错也要隐藏刷新状态
        hideRefreshingState();
    }
}


// 显示单封邮件
function displayEmail(mail) {
    const emailItem = document.createElement('div');
    emailItem.className = 'email-item';

    const now = new Date();
    const timeString = now.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });

    // 优先显示 HTML 内容，如果没有则显示文本内容
    let bodyContent = mail.HtmlContent || mail.TextContent;

    // 如果内容为空，显示提示
    if (!bodyContent) {
        bodyContent = '<p>(无内容)</p>';
    }

    emailItem.innerHTML = `
        <div class="email-item-header">
            <div class="email-item-from">${escapeHtml(mail.from || '未知发件人')}</div>
            <div class="email-item-date">${timeString}</div>
        </div>
        <div class="email-item-title">${escapeHtml(mail.title || '无主题')}</div>
        <div class="email-item-body">${bodyContent}</div>
    `;

    // 插入到邮件列表顶部
    emailContentEl.insertBefore(emailItem, emailContentEl.firstChild);
    
    // 更新邮件计数
    updateEmailCount();
}

// 更新邮件计数
function updateEmailCount() {
    const emailCount = emailContentEl.querySelectorAll('.email-item').length;
    emailCountEl.textContent = `${emailCount} 封邮件`;
}

// 复制邮箱地址到剪贴板
async function copyEmail() {
    if (!currentEmail) return;

    try {
        await navigator.clipboard.writeText(currentEmail);

        // 显示复制成功提示
        copyBtn.textContent = '✅ 已复制!';
        copyBtn.classList.add('copied');

        // 恢复原始状态
        setTimeout(() => {
            copyBtn.textContent = '📋 复制';
            copyBtn.classList.remove('copied');
        }, 2000);
    } catch (error) {
        console.error('复制失败:', error);

        // 降级方案：使用 input.select()
        emailAddressEl.select();
        document.execCommand('copy');

        // 显示复制成功提示
        copyBtn.textContent = '✅ 已复制!';
        copyBtn.classList.add('copied');

        // 恢复原始状态
        setTimeout(() => {
            copyBtn.textContent = '📋 复制';
            copyBtn.classList.remove('copied');
        }, 2000);
    }
}

// 开始自动刷新
function startAutoRefresh() {
    // 只有在非刷新状态才更新显示
    if (!isRefreshing) {
        countdownValue = 5;
        updateCountdown();
    }
    
    // 每5秒刷新一次
    refreshInterval = setInterval(refreshEmails, 5000);
    
    // 开始倒计时（只有在非刷新状态才开始新的倒计时）
    if (!countdownInterval) {
        countdownInterval = setInterval(() => {
            if (!isRefreshing) {
                countdownValue--;
                updateCountdown();
                if (countdownValue <= 0) {
                    countdownValue = 5;
                }
            }
        }, 1000);
    }
}

// 更新倒计时显示
function updateCountdown() {
    // 如果在刷新中，不更新倒计时
    if (isRefreshing) {
        return;
    }
    refreshCountdownEl.textContent = `${countdownValue}秒后刷新`;
}

// 显示刷新中状态
function showRefreshingState() {
    isRefreshing = true;
    refreshCountdownEl.textContent = '刷新中...';
    refreshCountdownEl.classList.add('refreshing');
}

// 隐藏刷新中状态，恢复正常倒计时
function hideRefreshingState() {
    isRefreshing = false;
    refreshCountdownEl.classList.remove('refreshing');
    // 重置倒计时值
    countdownValue = 5;
    updateCountdown();
}

// 停止自动刷新
function stopAutoRefresh() {
    if (refreshInterval) {
        clearInterval(refreshInterval);
        refreshInterval = null;
    }
    if (countdownInterval) {
        clearInterval(countdownInterval);
        countdownInterval = null;
    }
}

// 绑定事件监听器
function bindEvents() {
    // 模式切换
    randomModeBtn.addEventListener('click', () => switchMode('random'));
    customModeBtn.addEventListener('click', () => switchMode('custom'));

    // 生成新邮箱
    generateBtn.addEventListener('click', () => {
        generateEmail();
    });

    // 使用自定义邮箱
    useCustomBtn.addEventListener('click', useCustomEmail);

    // 复制邮箱地址
    copyBtn.addEventListener('click', copyEmail);

    // 手动刷新 - 点击倒计时后立即刷新
    refreshCountdownEl.addEventListener('click', () => {
        // 立即刷新，refreshEmails函数会处理刷新状态和倒计时
        refreshEmails();
    });

    // 点击邮箱地址也可以复制
    emailAddressEl.addEventListener('click', copyEmail);

    // 页面可见性变化时调整自动刷新
    document.addEventListener('visibilitychange', () => {
        if (document.hidden) {
            stopAutoRefresh();
        } else {
            startAutoRefresh();
            // 页面重新可见时立即刷新
            refreshEmails();
        }
    });

    // 回车键使用自定义邮箱
    customUsernameEl.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            useCustomEmail();
        }
    });
}

// 切换生成模式
function switchMode(mode) {
    // 更新按钮状态
    randomModeBtn.classList.toggle('active', mode === 'random');
    customModeBtn.classList.toggle('active', mode === 'custom');
    
    // 更新面板显示
    randomMode.classList.toggle('active', mode === 'random');
    customMode.classList.toggle('active', mode === 'custom');
    
    // 如果是随机模式且当前没有邮箱，自动生成一个
    if (mode === 'random' && !currentEmail) {
        generateEmail();
    }
}

// HTML 转义函数
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// 页面加载完成后初始化
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
} else {
    init();
}