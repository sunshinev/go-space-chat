<style scoped>
    .layout{
        border: 1px solid #d7dde4;
        background: #f5f7f9;
        position: relative;
        border-radius: 4px;
        overflow: hidden;
    }
    .layout-logo{
        width: 200px;
        height: 30px;
        font-size: 24px;
        border-radius: 3px;
        color: white;
        float: left;
        position: relative;
        top: 15px;
        left: 20px;
    }
    .layout-nav{
        width: 420px;
        margin: 0 auto;
        margin-right: 20px;
    }
    .layout-footer-center{
        text-align: center;
    }
</style>
<template>
    <Menu mode="horizontal" theme="dark" active-name="1">
        <div class="layout-logo">
            <div>Message board</div>
        </div>
        <div class="layout-nav">
            <MenuItem name="1">
                <Icon type="ios-navigate"></Icon>
                Item 1
            </MenuItem>
            <MenuItem name="2">
                <Icon type="ios-keypad"></Icon>
                Item 2
            </MenuItem>
            <MenuItem name="3">
                <Icon type="ios-analytics"></Icon>
                Item 3
            </MenuItem>
            <MenuItem name="4">
                <Icon type="ios-paper"></Icon>
                Item 4
            </MenuItem>
        </div>
    </Menu>
</template>
<script>
    export default {
        data () {
            return {
                theme1: 'light'
            }
        }
    }
</script>
