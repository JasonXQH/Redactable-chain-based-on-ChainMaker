pragma solidity ^0.8.0;


contract Devin{
    string[6] params;

    function cross_call(address callee, string memory  method, string memory to, string memory amount) internal  {
        //CrossVMCall is reserved key word
        params[0] = "CrossVMCall";
        params[1] = method;
        params[2] = "to";
        params[3] = to;
        params[4] = "amount";
        params[5] = amount;
        callee.call("");
    }

    function AddrToString(address account) public pure returns(string memory) {
        return toString(abi.encodePacked(account));
    }

    function IntToString(uint256 value) public pure returns(string memory) {
        return toString(abi.encodePacked(value));
    }

    function BytesToString(bytes32 value) public pure returns(string memory) {
        return toString(abi.encodePacked(value));
    }

    function toString(bytes memory data) public pure returns(string memory) {
        bytes memory alphabet = "0123456789abcdef";

        bytes memory str = new bytes(2 + data.length * 2);
        str[0] = "0";
        str[1] = "x";
        for (uint i = 0; i < data.length; i++) {
            str[2+i*2] = alphabet[uint(uint8(data[i] >> 4))];
            str[3+i*2] = alphabet[uint(uint8(data[i] & 0x0f))];
        }
        return string(str);
    }

    function withdraw(address _tokenid,uint256 _amount) external returns (bool){
        string memory trans = "transfer";
        address addr=msg.sender;
        cross_call(_tokenid,  trans, AddrToString(addr),IntToString( _amount));
        return true;
    }


}